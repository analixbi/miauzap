<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_Queue {
	const CRON_HOOK = 'mzw_process_queue';
	const ACTION_GROUP = 'miauzap-woocommerce';
	const WORKER_AJAX_ACTION = 'mzw_queue_worker';
	const WORKER_LOCK_KEY = 'mzw_queue_worker_running';
	const PROCESS_LOCK_KEY = 'mzw_queue_process_running';
	const WORKER_SPAWN_KEY = 'mzw_queue_worker_spawned';

	public static function table_name() {
		global $wpdb;
		return $wpdb->prefix . 'mzw_queue';
	}

	public static function create_tables() {
		global $wpdb;
		require_once ABSPATH . 'wp-admin/includes/upgrade.php';

		$table = self::table_name();
		$charset = $wpdb->get_charset_collate();

		$sql = "CREATE TABLE {$table} (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			type varchar(40) NOT NULL,
			related_id bigint(20) unsigned NOT NULL DEFAULT 0,
			phone varchar(40) NOT NULL,
			message longtext NOT NULL,
			instance_key varchar(100) NOT NULL DEFAULT '',
			status varchar(20) NOT NULL DEFAULT 'pending',
			scheduled_at datetime NOT NULL,
			attempts smallint(5) unsigned NOT NULL DEFAULT 0,
			dedupe_key varchar(191) NOT NULL DEFAULT '',
			last_error text NULL,
			response longtext NULL,
			sent_at datetime NULL,
			created_at datetime NOT NULL,
			PRIMARY KEY  (id),
			UNIQUE KEY dedupe_key (dedupe_key),
			KEY status_scheduled (status, scheduled_at),
			KEY instance_key (instance_key),
			KEY phone (phone),
			KEY related_id (related_id)
		) {$charset};";

		dbDelta($sql);
	}

	public static function enqueue($type, $phone, $message, $related_id = 0, $args = array()) {
		global $wpdb;

		$phone = MZW_WooCommerce::sanitize_phone_digits($phone);
		$message = trim((string) $message);
		if ('' === $phone || '' === $message) {
			return new WP_Error('mzw_invalid_queue_item', __('Telefone e mensagem sao obrigatorios.', 'miauzap-woocommerce'));
		}

		$dedupe_key = isset($args['dedupe_key']) ? sanitize_text_field($args['dedupe_key']) : '';
		if ('' === $dedupe_key) {
			$dedupe_key = md5($type . '|' . $related_id . '|' . $phone . '|' . $message . '|' . microtime(true));
		}

		$settings = MZW_Options::get_settings();
		$scheduled = isset($args['scheduled_at']) ? strtotime($args['scheduled_at']) : 0;
		if (!$scheduled) {
			$scheduled = self::next_scheduled_timestamp(!empty($args['otp']));
			$scheduled += absint($args['extra_delay'] ?? 0);
		}

		$data = array(
			'type'         => sanitize_key($type),
			'related_id'   => absint($related_id),
			'phone'        => $phone,
			'message'      => $message,
			'instance_key' => sanitize_key($args['instance_key'] ?? ''),
			'status'       => 'pending',
			'scheduled_at' => gmdate('Y-m-d H:i:s', $scheduled),
			'attempts'     => 0,
			'dedupe_key'   => $dedupe_key,
			'created_at'   => current_time('mysql', true),
		);

		$inserted = $wpdb->insert(self::table_name(), $data);
		if (false === $inserted) {
			if (false !== strpos((string) $wpdb->last_error, 'Duplicate')) {
				self::schedule_worker();
				return 0;
			}

			return new WP_Error('mzw_queue_insert_failed', $wpdb->last_error);
		}

		self::schedule_worker($scheduled);
		return (int) $wpdb->insert_id;
	}

	public static function schedule_worker($timestamp = null) {
		$timestamp = $timestamp ? (int) $timestamp : self::next_pending_timestamp();
		if (!$timestamp) {
			return;
		}

		$timestamp = max(time() + 5, $timestamp);

		if (function_exists('as_schedule_single_action')) {
			as_schedule_single_action($timestamp, self::CRON_HOOK, array(), self::ACTION_GROUP);
		} elseif (!wp_next_scheduled(self::CRON_HOOK)) {
			wp_schedule_single_event($timestamp, self::CRON_HOOK);
		}

		if ($timestamp <= time() + 65) {
			self::spawn_wp_cron();
		}

		self::spawn_background_worker($timestamp);
	}

	public static function maybe_dispatch_due() {
		if (defined('DOING_CRON') && DOING_CRON) {
			return;
		}

		if (!self::has_due_items()) {
			return;
		}

		self::schedule_worker(time() + 5);
	}

	private static function next_pending_timestamp() {
		global $wpdb;
		$next = $wpdb->get_var(
			"SELECT MIN(scheduled_at) FROM " . self::table_name() . " WHERE status = 'pending'"
		);

		return $next ? strtotime($next . ' UTC') : 0;
	}

	private static function has_due_items() {
		global $wpdb;
		return (bool) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT id FROM " . self::table_name() . " WHERE status = %s AND scheduled_at <= %s LIMIT 1",
				'pending',
				current_time('mysql', true)
			)
		);
	}

	private static function spawn_wp_cron() {
		if (defined('DISABLE_WP_CRON') && DISABLE_WP_CRON) {
			return;
		}

		$doing_wp_cron = sprintf('%.22F', microtime(true));
		wp_remote_post(
			site_url('wp-cron.php?doing_wp_cron=' . rawurlencode($doing_wp_cron)),
			array(
				'timeout'   => 1,
				'blocking'  => false,
				'sslverify' => apply_filters('https_local_ssl_verify', false),
			)
		);
	}

	private static function spawn_background_worker($timestamp) {
		if ((int) $timestamp > time() + (3 * MINUTE_IN_SECONDS)) {
			return;
		}
		if (get_transient(self::WORKER_LOCK_KEY) || get_transient(self::PROCESS_LOCK_KEY) || get_transient(self::WORKER_SPAWN_KEY)) {
			return;
		}

		set_transient(self::WORKER_SPAWN_KEY, 1, 15);
		wp_remote_post(
			admin_url('admin-ajax.php'),
			array(
				'timeout'   => 1,
				'blocking'  => false,
				'sslverify' => apply_filters('https_local_ssl_verify', false),
				'body'      => array(
					'action' => self::WORKER_AJAX_ACTION,
					'key'    => wp_hash(self::WORKER_AJAX_ACTION),
				),
			)
		);
	}

	public static function handle_background_worker() {
		$key = sanitize_text_field(wp_unslash($_POST['key'] ?? ''));
		if (!hash_equals(wp_hash(self::WORKER_AJAX_ACTION), $key)) {
			wp_die('forbidden', '', array('response' => 403));
		}
		if (get_transient(self::WORKER_LOCK_KEY) || get_transient(self::PROCESS_LOCK_KEY)) {
			wp_die('locked');
		}

		delete_transient(self::WORKER_SPAWN_KEY);
		set_transient(self::WORKER_LOCK_KEY, 1, 3 * MINUTE_IN_SECONDS);

		if (function_exists('ignore_user_abort')) {
			ignore_user_abort(true);
		}
		if (function_exists('set_time_limit')) {
			@set_time_limit(180);
		}

		$settings = MZW_Options::get_settings();
		$deadline = time() + 150;
		$sent = 0;
		$max_per_run = max(1, absint($settings['process_limit']));
		while (time() < $deadline) {
			$next = self::next_pending_timestamp();
			if (!$next) {
				break;
			}

			$wait = $next - time();
			if ($wait > 0) {
				if ($wait > 20) {
					break;
				}
				sleep(min(5, $wait));
				continue;
			}

			self::process_due(1);
			$sent++;
			if ($sent >= $max_per_run) {
				break;
			}
			usleep(250000);
		}

		delete_transient(self::WORKER_LOCK_KEY);
		self::schedule_worker();
		wp_die('ok');
	}

	private static function next_scheduled_timestamp($otp = false) {
		global $wpdb;
		$settings = MZW_Options::get_settings();

		$min = $otp ? absint($settings['otp_min_delay']) : absint($settings['min_delay']);
		$max = $otp ? absint($settings['otp_max_delay']) : absint($settings['max_delay']);
		if ($max < $min) {
			$max = $min;
		}

		$last = $wpdb->get_var(
			"SELECT MAX(scheduled_at) FROM " . self::table_name() . " WHERE status = 'pending'"
		);
		$base = time();
		if ($last) {
			$last_ts = strtotime($last . ' UTC');
			$base = max($base, $last_ts);
		}

		return $base + random_int($min, $max);
	}

	public static function process_due($limit = null) {
		global $wpdb;
		$owns_lock = false;

		if (!self::is_background_worker_request()) {
			if (get_transient(self::WORKER_LOCK_KEY) || get_transient(self::PROCESS_LOCK_KEY)) {
				self::schedule_worker();
				return;
			}
			set_transient(self::PROCESS_LOCK_KEY, 1, MINUTE_IN_SECONDS);
			$owns_lock = true;
		}

		$items = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT * FROM " . self::table_name() . " WHERE status = %s AND scheduled_at <= %s ORDER BY scheduled_at ASC, id ASC LIMIT %d",
				'pending',
				current_time('mysql', true),
				1
			),
			ARRAY_A
		);

		foreach ($items as $item) {
			self::send_item($item);
		}

		self::spread_due_items();
		if ($owns_lock) {
			delete_transient(self::PROCESS_LOCK_KEY);
		}
		self::schedule_worker();
	}

	private static function is_background_worker_request() {
		return defined('DOING_AJAX') && DOING_AJAX && isset($_POST['action']) && self::WORKER_AJAX_ACTION === sanitize_key(wp_unslash($_POST['action']));
	}

	private static function spread_due_items() {
		global $wpdb;
		$settings = MZW_Options::get_settings();

		$items = $wpdb->get_results(
			$wpdb->prepare(
				"SELECT id FROM " . self::table_name() . " WHERE status = %s AND scheduled_at <= %s ORDER BY scheduled_at ASC, id ASC",
				'pending',
				current_time('mysql', true)
			),
			ARRAY_A
		);
		if (empty($items)) {
			return;
		}

		$min = absint($settings['min_delay']);
		$max = absint($settings['max_delay']);
		if ($max < $min) {
			$max = $min;
		}
		$base = time();

		foreach ($items as $item) {
			$base += max(1, random_int($min, $max));
			$wpdb->update(
				self::table_name(),
				array('scheduled_at' => gmdate('Y-m-d H:i:s', $base)),
				array('id' => absint($item['id']))
			);
		}
	}

	private static function send_item($item) {
		global $wpdb;
		$settings = MZW_Options::get_settings();

		if (self::is_quiet_hours()) {
			self::reschedule($item['id'], self::next_open_timestamp());
			return;
		}

		$instance = null;
		if (!empty($item['instance_key'])) {
			$instance = MZW_Options::get_instance($item['instance_key']);
		}
		if (!$instance) {
			$instance = MZW_Options::pick_instance();
		}
		if (!$instance) {
			self::fail($item, __('Nenhuma instancia Miauzap ativa foi configurada.', 'miauzap-woocommerce'), true);
			return;
		}

		if (self::daily_limit_reached($instance['key'])) {
			self::reschedule($item['id'], time() + HOUR_IN_SECONDS);
			return;
		}

		$claimed = $wpdb->update(
			self::table_name(),
			array('status' => 'processing', 'instance_key' => $instance['key']),
			array('id' => absint($item['id']), 'status' => 'pending')
		);
		if (false === $claimed || 0 === $claimed) {
			return;
		}

		$client = new MZW_API_Client();

		$send_phone = MZW_WooCommerce::normalize_phone($item['phone']);
		$resolved_phone = self::resolve_whatsapp_phone($client, $instance, $item['phone']);
		if (is_wp_error($resolved_phone)) {
			if (!empty($settings['check_whatsapp'])) {
				self::fail($item, $resolved_phone->get_error_message(), false);
				return;
			}
			$resolved_phone = false;
		}
		if ($resolved_phone) {
			$send_phone = $resolved_phone;
		} elseif (!empty($settings['check_whatsapp'])) {
			self::fail($item, __('Numero nao encontrado no WhatsApp.', 'miauzap-woocommerce'), true, 'skipped');
			return;
		}

		$response = $client->send_text(
			$instance,
			$send_phone,
			$item['message'],
			array('link_preview' => !empty($settings['link_preview']))
		);

		if (is_wp_error($response)) {
			self::fail($item, $response->get_error_message(), false);
			return;
		}

		$wpdb->update(
			self::table_name(),
			array(
				'status'   => 'sent',
				'response' => wp_json_encode($response),
				'sent_at'  => current_time('mysql', true),
			),
			array('id' => absint($item['id']))
		);
	}

	private static function fail($item, $message, $final = false, $status = 'failed') {
		global $wpdb;
		$settings = MZW_Options::get_settings();
		$attempts = absint($item['attempts']) + 1;
		$max = max(1, absint($settings['max_attempts']));

		if (!$final && $attempts < $max) {
			$wpdb->update(
				self::table_name(),
				array(
					'status'       => 'pending',
					'attempts'     => $attempts,
					'last_error'   => $message,
					'scheduled_at' => gmdate('Y-m-d H:i:s', time() + (MINUTE_IN_SECONDS * $attempts)),
				),
				array('id' => absint($item['id']))
			);
			self::schedule_worker(time() + (MINUTE_IN_SECONDS * $attempts));
			return;
		}

		$wpdb->update(
			self::table_name(),
			array(
				'status'     => $status,
				'attempts'   => $attempts,
				'last_error' => $message,
			),
			array('id' => absint($item['id']))
		);
	}

	private static function reschedule($id, $timestamp) {
		global $wpdb;
		$wpdb->update(
			self::table_name(),
			array(
				'status'       => 'pending',
				'scheduled_at' => gmdate('Y-m-d H:i:s', $timestamp),
			),
			array('id' => absint($id))
		);
		self::schedule_worker($timestamp);
	}

	private static function resolve_whatsapp_phone($client, $instance, $phone) {
		$last_error = null;
		foreach (MZW_WooCommerce::phone_variants($phone) as $candidate) {
			$contact = $client->get_contact($instance, $candidate);
			if (is_wp_error($contact)) {
				$last_error = $contact;
				continue;
			}
			if ($contact) {
				if (!empty($contact['jid'])) {
					return $contact['jid'];
				}
				if (!empty($contact['phone'])) {
					return $contact['phone'];
				}
				return $candidate;
			}
		}

		return $last_error ? $last_error : false;
	}

	private static function daily_limit_reached($instance_key) {
		global $wpdb;
		$settings = MZW_Options::get_settings();
		$limit = absint($settings['daily_limit_per_instance']);
		if (!$limit) {
			return false;
		}

		$count = (int) $wpdb->get_var(
			$wpdb->prepare(
				"SELECT COUNT(*) FROM " . self::table_name() . " WHERE status = 'sent' AND instance_key = %s AND sent_at >= %s",
				$instance_key,
				gmdate('Y-m-d H:i:s', time() - DAY_IN_SECONDS)
			)
		);

		return $count >= $limit;
	}

	private static function is_quiet_hours() {
		$settings = MZW_Options::get_settings();
		if (empty($settings['quiet_hours_enabled'])) {
			return false;
		}

		$now = (int) current_time('Hi');
		$start = (int) str_replace(':', '', $settings['quiet_hours_start']);
		$end = (int) str_replace(':', '', $settings['quiet_hours_end']);

		if ($start === $end) {
			return false;
		}

		if ($start < $end) {
			return $now >= $start && $now < $end;
		}

		return $now >= $start || $now < $end;
	}

	private static function next_open_timestamp() {
		$settings = MZW_Options::get_settings();
		$end = preg_match('/^\d{2}:\d{2}$/', $settings['quiet_hours_end']) ? $settings['quiet_hours_end'] : '08:00';
		$timezone = function_exists('wp_timezone') ? wp_timezone() : new DateTimeZone('UTC');
		$now = new DateTimeImmutable('now', $timezone);
		$target = new DateTimeImmutable($now->format('Y-m-d') . ' ' . $end, $timezone);

		if ($target->getTimestamp() <= $now->getTimestamp()) {
			$target = $target->modify('+1 day');
		}

		return $target->getTimestamp() + random_int(absint($settings['min_delay']), max(absint($settings['min_delay']), absint($settings['max_delay'])));
	}

	public static function retry($id) {
		global $wpdb;
		$result = $wpdb->update(
			self::table_name(),
			array(
				'status'       => 'pending',
				'attempts'     => 0,
				'last_error'   => '',
				'scheduled_at' => current_time('mysql', true),
			),
			array('id' => absint($id))
		);
		self::schedule_worker(time() + 5);
		return $result;
	}

	public static function get_items($status = '', $limit = 50) {
		global $wpdb;
		$where = '';
		$args = array();
		if ($status) {
			$where = 'WHERE status = %s';
			$args[] = sanitize_key($status);
		}

		$sql = "SELECT * FROM " . self::table_name() . " {$where} ORDER BY id DESC LIMIT %d";
		$args[] = absint($limit);

		return $wpdb->get_results($wpdb->prepare($sql, $args), ARRAY_A);
	}

	public static function stats() {
		global $wpdb;
		$rows = $wpdb->get_results("SELECT status, COUNT(*) qty FROM " . self::table_name() . " GROUP BY status", ARRAY_A);
		$out = array('pending' => 0, 'processing' => 0, 'sent' => 0, 'failed' => 0, 'skipped' => 0);
		foreach ($rows as $row) {
			$out[$row['status']] = absint($row['qty']);
		}

		return $out;
	}
}
