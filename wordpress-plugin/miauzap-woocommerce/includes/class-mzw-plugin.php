<?php

if (!defined('ABSPATH')) {
	exit;
}

final class MZW_Plugin {
	private static $instance = null;

	public static function instance() {
		if (null === self::$instance) {
			self::$instance = new self();
		}

		return self::$instance;
	}

	private function __construct() {
		add_filter('cron_schedules', array($this, 'cron_schedules'));
		add_action(MZW_Queue::CRON_HOOK, array('MZW_Queue', 'process_due'));
		add_action('wp_ajax_nopriv_' . MZW_Queue::WORKER_AJAX_ACTION, array('MZW_Queue', 'handle_background_worker'));
		add_action('wp_ajax_' . MZW_Queue::WORKER_AJAX_ACTION, array('MZW_Queue', 'handle_background_worker'));
		add_action('init', array('MZW_Queue', 'maybe_dispatch_due'), 20);

		new MZW_Admin();
		new MZW_OTP();

		if (class_exists('WooCommerce')) {
			new MZW_WooCommerce();
		} else {
			add_action('admin_notices', array($this, 'woocommerce_notice'));
		}
	}

	public static function activate() {
		MZW_Queue::create_tables();
		MZW_Options::install_defaults();

		if (!wp_next_scheduled(MZW_Queue::CRON_HOOK)) {
			wp_schedule_event(time() + 60, 'mzw_every_minute', MZW_Queue::CRON_HOOK);
		}
		MZW_Queue::schedule_worker(time() + 60);
	}

	public static function deactivate() {
		wp_clear_scheduled_hook(MZW_Queue::CRON_HOOK);
	}

	public function cron_schedules($schedules) {
		$schedules['mzw_every_minute'] = array(
			'interval' => 60,
			'display'  => __('A cada minuto', 'miauzap-woocommerce'),
		);

		return $schedules;
	}

	public function woocommerce_notice() {
		if (!current_user_can('activate_plugins')) {
			return;
		}

		echo '<div class="notice notice-warning"><p>';
		echo esc_html__('Miauzap para WooCommerce esta ativo, mas o WooCommerce nao foi encontrado. O OTP e o cliente da API ainda funcionam, mas automacoes de pedidos e produtos precisam do WooCommerce.', 'miauzap-woocommerce');
		echo '</p></div>';
	}
}
