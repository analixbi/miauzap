<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_API_Client {
	private $base_url;

	public function __construct($base_url = '') {
		$settings = MZW_Options::get_settings();
		$this->base_url = rtrim($base_url ? $base_url : $settings['base_url'], '/');
	}

	public function send_text($instance, $phone, $message, $args = array()) {
		$message = $this->prepare_message($message);
		$payload = array(
			'number'      => $phone,
			'text'        => $message,
			'linkPreview' => !empty($args['link_preview']),
		);

		if (!empty($args['id'])) {
			$payload['id'] = $args['id'];
		}

		return $this->request('POST', '/chat/send/text', $instance, $payload);
	}

	private function prepare_message($message) {
		$charset = function_exists('get_bloginfo') ? get_bloginfo('charset') : 'UTF-8';
		$text = html_entity_decode((string) $message, ENT_QUOTES | ENT_HTML5, $charset ? $charset : 'UTF-8');
		$text = str_replace(array('&nbsp;', '&#160;', '&#xA0;', "\xc2\xa0"), ' ', $text);

		return $text;
	}

	public function get_contact($instance, $phone) {
		$phone = trim((string) $phone);
		if ('' === $phone) {
			return false;
		}

		$response = $this->request('POST', '/user/check', $instance, array('Phone' => array($phone)));
		if (is_wp_error($response)) {
			return $response;
		}

		$data = $this->unwrap_data($response);
		$users = array();
		if (isset($data['Users']) && is_array($data['Users'])) {
			$users = $data['Users'];
		} elseif (isset($data['users']) && is_array($data['users'])) {
			$users = $data['users'];
		}

		foreach ($users as $user) {
			if (!is_array($user)) {
				continue;
			}

			$is_in_whatsapp = false;
			foreach (array('IsInWhatsapp', 'isInWhatsapp', 'is_in_whatsapp') as $key) {
				if (array_key_exists($key, $user)) {
					$value = $user[$key];
					$is_in_whatsapp = is_bool($value) ? $value : in_array(strtolower((string) $value), array('1', 'true', 'yes', 'sim'), true);
					break;
				}
			}

			if (!$is_in_whatsapp) {
				continue;
			}

			$jid = '';
			foreach (array('JID', 'jid', 'Jid') as $key) {
				if (!empty($user[$key])) {
					$jid = sanitize_text_field((string) $user[$key]);
					break;
				}
			}

			$resolved_phone = '';
			foreach (array('Phone', 'phone', 'Number', 'number', 'Query', 'query') as $key) {
				if (!empty($user[$key])) {
					$resolved_phone = sanitize_text_field((string) $user[$key]);
					break;
				}
			}

			return array(
				'query' => sanitize_text_field($phone),
				'phone' => $resolved_phone ? $resolved_phone : sanitize_text_field($phone),
				'jid'   => $jid,
				'raw'   => $user,
			);
		}

		return false;
	}

	public function check_number($instance, $phone) {
		$contact = $this->get_contact($instance, $phone);
		return is_wp_error($contact) ? $contact : (bool) $contact;
	}

	public function status($instance) {
		return $this->request('GET', '/session/status', $instance);
	}

	public function health() {
		return $this->request('GET', '/health', null);
	}

	public function request($method, $path, $instance = null, $payload = null) {
		if (empty($this->base_url)) {
			return new WP_Error('mzw_missing_base_url', __('A URL base do Miauzap esta vazia.', 'miauzap-woocommerce'));
		}

		$url = $this->base_url . '/' . ltrim($path, '/');
		$args = array(
			'timeout' => 25,
			'headers' => array(
				'Accept'       => 'application/json',
				'Content-Type' => 'application/json',
				'User-Agent'   => 'Miauzap-WooCommerce/' . MZW_VERSION,
			),
		);

		if (is_array($instance) && !empty($instance['token'])) {
			$args['headers']['Authorization'] = $instance['token'];
			$args['headers']['Token'] = $instance['token'];
		}

		if (null !== $payload) {
			$args['body'] = wp_json_encode($payload);
		}

		$response = 'GET' === strtoupper($method) ? wp_remote_get($url, $args) : wp_remote_post($url, $args);
		if (is_wp_error($response)) {
			return $response;
		}

		$code = (int) wp_remote_retrieve_response_code($response);
		$body = wp_remote_retrieve_body($response);
		$decoded = json_decode($body, true);
		$result = null === $decoded ? array('raw' => $body) : $decoded;

		if ($code < 200 || $code >= 300) {
			return new WP_Error('mzw_api_error', sprintf(__('A API Miauzap retornou HTTP %d.', 'miauzap-woocommerce'), $code), $result);
		}

		return $result;
	}

	public function unwrap_data($response) {
		if (isset($response['data'])) {
			if (is_string($response['data'])) {
				$decoded = json_decode($response['data'], true);
				return is_array($decoded) ? $decoded : array('raw' => $response['data']);
			}

			return is_array($response['data']) ? $response['data'] : array();
		}

		return is_array($response) ? $response : array();
	}
}
