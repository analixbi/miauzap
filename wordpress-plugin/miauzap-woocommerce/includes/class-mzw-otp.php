<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_OTP {
	public function __construct() {
		add_shortcode('miauzap_otp_login', array($this, 'shortcode'));
		add_action('woocommerce_login_form_end', array($this, 'login_form'));
		add_action('wp_enqueue_scripts', array($this, 'enqueue_assets'));

		add_action('wp_ajax_nopriv_mzw_send_otp', array($this, 'ajax_send_otp'));
		add_action('wp_ajax_mzw_send_otp', array($this, 'ajax_send_otp'));
		add_action('wp_ajax_nopriv_mzw_verify_otp', array($this, 'ajax_verify_otp'));
		add_action('wp_ajax_mzw_verify_otp', array($this, 'ajax_verify_otp'));
	}

	public function enqueue_assets() {
		wp_enqueue_style('mzw-frontend', MZW_URL . 'assets/frontend.css', array(), MZW_VERSION);
		wp_enqueue_script('mzw-frontend', MZW_URL . 'assets/frontend.js', array('jquery'), MZW_VERSION, true);
		wp_localize_script(
			'mzw-frontend',
			'MZWFrontend',
			array(
				'ajaxUrl' => admin_url('admin-ajax.php'),
				'nonce'   => wp_create_nonce('mzw_otp'),
			)
		);
	}

	public function login_form() {
		echo $this->shortcode();
	}

	public function shortcode() {
		$settings = MZW_Options::get_settings();
		if (empty($settings['otp_enabled'])) {
			return '';
		}

		ob_start();
		?>
		<div class="mzw-otp-box">
			<h3><?php echo esc_html__('Entrar com WhatsApp', 'miauzap-woocommerce'); ?></h3>
			<p class="mzw-otp-message" aria-live="polite"></p>
			<p>
				<label><?php echo esc_html__('DDD e numero', 'miauzap-woocommerce'); ?></label>
				<input type="tel" class="mzw-otp-phone" inputmode="numeric" maxlength="15" autocomplete="tel-national" placeholder="11999999999">
			</p>
			<p>
				<button type="button" class="button mzw-send-otp mzw-whatsapp-button">
					<span class="mzw-whatsapp-icon" aria-hidden="true">
						<svg viewBox="0 0 32 32" focusable="false">
							<path d="M16.01 3.2c-7.05 0-12.78 5.67-12.78 12.65 0 2.23.59 4.42 1.7 6.34l-1.8 6.61 6.84-1.77a12.9 12.9 0 0 0 6.04 1.51c7.04 0 12.77-5.68 12.77-12.66S23.05 3.2 16.01 3.2Zm0 22.98c-1.95 0-3.85-.53-5.51-1.54l-.39-.23-4.06 1.05 1.08-3.92-.25-.4a10.16 10.16 0 0 1-1.57-5.29c0-5.67 4.8-10.29 10.7-10.29 5.89 0 10.68 4.62 10.68 10.32 0 5.68-4.79 10.3-10.68 10.3Zm5.86-7.72c-.32-.16-1.9-.93-2.19-1.04-.29-.11-.5-.16-.71.16-.21.31-.82 1.03-1 1.24-.18.21-.37.24-.69.08-.32-.16-1.35-.49-2.57-1.57-.95-.84-1.59-1.88-1.78-2.2-.18-.31-.02-.48.14-.64.14-.14.32-.37.48-.56.16-.18.21-.31.32-.53.11-.21.05-.4-.03-.56-.08-.16-.71-1.69-.98-2.32-.26-.6-.52-.52-.71-.53h-.61c-.21 0-.56.08-.85.4-.29.31-1.11 1.07-1.11 2.61 0 1.53 1.13 3.02 1.29 3.23.16.21 2.23 3.36 5.4 4.72.75.32 1.34.51 1.8.65.76.24 1.45.2 2 .12.61-.09 1.9-.77 2.17-1.51.27-.75.27-1.39.19-1.52-.08-.13-.29-.21-.61-.37Z" />
						</svg>
					</span>
					<span><?php echo esc_html__('Enviar codigo', 'miauzap-woocommerce'); ?></span>
				</button>
			</p>
			<p class="mzw-otp-code-wrap" hidden>
				<label><?php echo esc_html__('Codigo recebido no WhatsApp', 'miauzap-woocommerce'); ?></label>
				<input type="text" inputmode="numeric" class="mzw-otp-code" maxlength="8">
				<button type="button" class="button button-primary mzw-verify-otp"><?php echo esc_html__('Entrar', 'miauzap-woocommerce'); ?></button>
			</p>
		</div>
		<?php
		return ob_get_clean();
	}

	public function ajax_send_otp() {
		check_ajax_referer('mzw_otp', 'nonce');

		$settings = MZW_Options::get_settings();
		if (empty($settings['otp_enabled'])) {
			wp_send_json_error(array('message' => __('Login por WhatsApp esta desativado.', 'miauzap-woocommerce')), 403);
		}

		$phone = $this->normalize_login_phone(wp_unslash($_POST['phone'] ?? ''));
		if (!$phone) {
			wp_send_json_error(array('message' => __('Informe DDD e numero do telefone.', 'miauzap-woocommerce')), 400);
		}
		$login_phone = $phone;
		$send_phone = $phone;

		$existing_user = $this->find_user_by_phone($phone);
		if (!$existing_user) {
			wp_send_json_error(array('message' => __('Conta nao encontrada para este telefone.', 'miauzap-woocommerce')), 404);
		}

		$rate_key = 'mzw_otp_rate_' . md5($phone . '|' . $this->request_ip());
		if (get_transient($rate_key)) {
			wp_send_json_error(array('message' => __('Aguarde um instante antes de pedir outro codigo.', 'miauzap-woocommerce')), 429);
		}

		$instance = MZW_Options::pick_instance();
		if (!$instance) {
			wp_send_json_error(array('message' => __('Nenhuma instancia Miauzap configurada.', 'miauzap-woocommerce')), 500);
		}

		$client = new MZW_API_Client();
		$resolved_phone = $this->resolve_whatsapp_phone($client, $instance, $phone);
		if (is_wp_error($resolved_phone)) {
			if (!empty($settings['check_whatsapp'])) {
				wp_send_json_error(array('message' => $resolved_phone->get_error_message()), 500);
			}
			$resolved_phone = false;
		}
		if ($resolved_phone) {
			$send_phone = $resolved_phone;
		} elseif (!empty($settings['check_whatsapp'])) {
			wp_send_json_error(array('message' => __('Este telefone nao foi encontrado no WhatsApp.', 'miauzap-woocommerce')), 400);
		}

		$length = min(8, max(4, absint($settings['otp_length'])));
		$min = (int) pow(10, $length - 1);
		$max = (int) pow(10, $length) - 1;
		$code = (string) random_int($min, $max);
		$ttl = max(1, absint($settings['otp_ttl_minutes']));
		$message = MZW_WooCommerce::render_template(
			$settings['otp_template'],
			array(
				'{otp}'       => $code,
				'{site_name}' => get_bloginfo('name'),
				'{site_url}'  => home_url('/'),
				'{minutes}'   => $ttl,
			)
		);

		$delay = random_int(absint($settings['otp_min_delay']), max(absint($settings['otp_min_delay']), absint($settings['otp_max_delay'])));
		if ($delay > 0) {
			sleep($delay);
		}

		$response = $client->send_text($instance, $send_phone, $message, array('link_preview' => false));
		if (is_wp_error($response)) {
			wp_send_json_error(array('message' => $response->get_error_message()), 500);
		}

		set_transient($this->otp_key($login_phone), array(
			'hash'    => wp_hash_password($code),
			'phone'   => $login_phone,
			'user_id' => $existing_user->ID,
		), $ttl * MINUTE_IN_SECONDS);
		set_transient($rate_key, 1, MINUTE_IN_SECONDS);

		wp_send_json_success(array('message' => __('Codigo enviado pelo WhatsApp.', 'miauzap-woocommerce')));
	}

	public function ajax_verify_otp() {
		check_ajax_referer('mzw_otp', 'nonce');

		$phone = $this->normalize_login_phone(wp_unslash($_POST['phone'] ?? ''));
		$code = preg_replace('/\D+/', '', wp_unslash($_POST['code'] ?? ''));

		$data = get_transient($this->otp_key($phone));
		if (!$phone || !$code || !is_array($data) || empty($data['hash']) || !wp_check_password($code, $data['hash'])) {
			wp_send_json_error(array('message' => __('Codigo invalido ou expirado.', 'miauzap-woocommerce')), 400);
		}

		$user = !empty($data['user_id']) ? get_user_by('id', absint($data['user_id'])) : $this->find_user_by_phone($phone);

		if (is_wp_error($user)) {
			wp_send_json_error(array('message' => $user->get_error_message()), 500);
		}
		if (!$user) {
			wp_send_json_error(array('message' => __('Conta nao encontrada para este telefone.', 'miauzap-woocommerce')), 404);
		}

		delete_transient($this->otp_key($phone));
		wp_set_current_user($user->ID);
		wp_set_auth_cookie($user->ID, true);

		wp_send_json_success(array(
			'message'  => __('Login concluido.', 'miauzap-woocommerce'),
			'redirect' => function_exists('wc_get_page_permalink') ? wc_get_page_permalink('myaccount') : home_url('/'),
		));
	}

	private function find_user_by_phone($phone) {
		$variants = $this->login_phone_variants($phone);
		if (empty($variants)) {
			return null;
		}

		$query = new WP_User_Query(array(
			'number'     => 1,
			'meta_query' => array(
				'relation' => 'OR',
				array('key' => 'billing_phone', 'value' => $variants, 'compare' => 'IN'),
			),
		));

		$users = $query->get_results();
		if (!empty($users)) {
			return $users[0];
		}

		$tail = substr($this->normalize_login_phone($phone), -4);
		$fallback = new WP_User_Query(array(
			'number'     => 100,
			'meta_query' => array(
				array('key' => 'billing_phone', 'value' => $tail, 'compare' => 'LIKE'),
			),
		));

		foreach ($fallback->get_results() as $user) {
			$saved_phone = get_user_meta($user->ID, 'billing_phone', true);
			if (array_intersect($variants, $this->login_phone_variants($saved_phone))) {
				MZW_WooCommerce::save_user_phone_meta($user->ID, $saved_phone);
				return $user;
			}
		}

		return null;
	}

	private function resolve_whatsapp_phone($client, $instance, $phone) {
		$last_error = null;
		foreach ($this->login_phone_variants($phone) as $candidate) {
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

	private function normalize_login_phone($phone) {
		$digits = preg_replace('/\D+/', '', (string) $phone);
		if (0 === strpos($digits, '55') && in_array(strlen($digits), array(12, 13), true)) {
			$digits = substr($digits, 2);
		}

		return preg_match('/^\d{10,11}$/', $digits) ? $digits : '';
	}

	private function login_phone_variants($phone) {
		$digits = $this->normalize_login_phone($phone);
		if (!$digits) {
			return array();
		}

		$settings = MZW_Options::get_settings();
		$country = preg_replace('/\D+/', '', (string) $settings['default_country_code']);
		$country = $country ? $country : '55';
		$variants = array();

		if (10 === strlen($digits)) {
			$with_nine = substr($digits, 0, 2) . '9' . substr($digits, 2);
			$variants = array(
				$country . $with_nine,
				$country . $digits,
				$with_nine,
				$digits,
			);
		} elseif (11 === strlen($digits) && '9' === substr($digits, 2, 1)) {
			$without_nine = substr($digits, 0, 2) . substr($digits, 3);
			$variants = array(
				$country . $digits,
				$country . $without_nine,
				$digits,
				$without_nine,
			);
		} else {
			$variants = array(
				$country . $digits,
				$digits,
			);
		}

		return array_values(array_unique(array_filter($variants)));
	}

	private function otp_key($phone) {
		return 'mzw_otp_' . md5($phone . '|' . wp_salt('auth'));
	}

	private function request_ip() {
		return sanitize_text_field(wp_unslash($_SERVER['REMOTE_ADDR'] ?? ''));
	}
}
