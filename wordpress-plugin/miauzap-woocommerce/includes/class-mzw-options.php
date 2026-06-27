<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_Options {
	const SETTINGS_KEY = 'mzw_settings';
	const ORDER_TEMPLATES_KEY = 'mzw_order_templates';
	const MARKETING_KEY = 'mzw_marketing';

	public static function install_defaults() {
		if (false === get_option(self::SETTINGS_KEY, false)) {
			add_option(self::SETTINGS_KEY, self::defaults(), '', false);
		}

		if (false === get_option(self::ORDER_TEMPLATES_KEY, false)) {
			add_option(self::ORDER_TEMPLATES_KEY, array(), '', false);
		}

		if (false === get_option(self::MARKETING_KEY, false)) {
			add_option(self::MARKETING_KEY, self::marketing_defaults(), '', false);
		}
	}

	public static function defaults() {
		return array(
			'base_url'                  => '',
			'default_country_code'      => '55',
			'instances'                 => array(),
			'check_whatsapp'            => 1,
			'link_preview'              => 1,
			'min_delay'                 => 10,
			'max_delay'                 => 30,
			'otp_min_delay'             => 0,
			'otp_max_delay'             => 2,
			'process_limit'             => 5,
			'max_attempts'              => 3,
			'daily_limit_per_instance'  => 250,
			'quiet_hours_enabled'       => 0,
			'quiet_hours_start'         => '22:00',
			'quiet_hours_end'           => '08:00',
			'marketing_require_optin'   => 1,
			'marketing_recipient_limit' => 200,
			'otp_enabled'               => 1,
			'otp_create_customer'       => 0,
			'otp_length'                => 6,
			'otp_ttl_minutes'           => 10,
			'otp_template'              => "Seu codigo de acesso da {site_name}: {otp}\nValido por {minutes} minutos.",
		);
	}

	public static function marketing_defaults() {
		return array(
			'new_product_enabled' => 0,
			'new_product_message' => "Novidade na {nome_site}: {nome_produto}\n{link_produto}",
			'stock_enabled'       => 0,
			'stock_message'       => "{nome_produto} voltou ao estoque.\nAcesse: {link_produto}",
		);
	}

	public static function get_settings() {
		$settings = get_option(self::SETTINGS_KEY, array());
		return wp_parse_args(is_array($settings) ? $settings : array(), self::defaults());
	}

	public static function update_settings($settings) {
		update_option(self::SETTINGS_KEY, wp_parse_args($settings, self::defaults()), false);
	}

	public static function get_order_templates() {
		$templates = get_option(self::ORDER_TEMPLATES_KEY, array());
		return is_array($templates) ? $templates : array();
	}

	public static function update_order_templates($templates) {
		update_option(self::ORDER_TEMPLATES_KEY, is_array($templates) ? $templates : array(), false);
	}

	public static function get_marketing() {
		$marketing = get_option(self::MARKETING_KEY, array());
		return wp_parse_args(is_array($marketing) ? $marketing : array(), self::marketing_defaults());
	}

	public static function update_marketing($marketing) {
		update_option(self::MARKETING_KEY, wp_parse_args($marketing, self::marketing_defaults()), false);
	}

	public static function get_instances($enabled_only = true) {
		$settings = self::get_settings();
		$instances = isset($settings['instances']) && is_array($settings['instances']) ? $settings['instances'] : array();
		$out = array();

		foreach ($instances as $key => $instance) {
			if (!is_array($instance)) {
				continue;
			}

			$token = isset($instance['token']) ? trim((string) $instance['token']) : '';
			if ('' === $token) {
				continue;
			}

			$enabled = !empty($instance['enabled']);
			if ($enabled_only && !$enabled) {
				continue;
			}

			$instance['key'] = is_string($key) ? $key : sanitize_key($instance['label'] ?? ('instance_' . $key));
			$instance['weight'] = max(1, absint($instance['weight'] ?? 1));
			$instance['label'] = sanitize_text_field($instance['label'] ?? $instance['key']);
			$out[$instance['key']] = $instance;
		}

		return $out;
	}

	public static function get_instance($key) {
		$instances = self::get_instances(false);
		return isset($instances[$key]) ? $instances[$key] : null;
	}

	public static function pick_instance() {
		$instances = self::get_instances(true);
		if (empty($instances)) {
			return null;
		}

		$total = 0;
		foreach ($instances as $instance) {
			$total += max(1, absint($instance['weight']));
		}

		$target = random_int(1, max(1, $total));
		$cursor = 0;
		foreach ($instances as $instance) {
			$cursor += max(1, absint($instance['weight']));
			if ($target <= $cursor) {
				return $instance;
			}
		}

		return reset($instances);
	}

	public static function order_variables() {
		return array(
			'{pedido_id}',
			'{numero_pedido}',
			'{order_id}',
			'{order_number}',
			'{status}',
			'{nome_cliente}',
			'{primeiro_nome}',
			'{sobrenome}',
			'{customer_name}',
			'{first_name}',
			'{last_name}',
			'{total}',
			'{moeda}',
			'{pagamento}',
			'{entrega}',
			'{itens}',
			'{currency}',
			'{payment_method}',
			'{shipping_method}',
			'{items}',
			'{telefone}',
			'{email}',
			'{data_pedido}',
			'{link_pedido}',
			'{correios_tracking_code}',
			'{codigo_rastreio}',
			'{billing_phone}',
			'{billing_email}',
			'{order_date}',
			'{order_url}',
			'{nome_site}',
			'{url_site}',
			'{site_name}',
			'{site_url}',
			'{nota:palavra_chave}',
			'{nota_cliente:palavra_chave}',
			'{nota_privada:palavra_chave}',
			'{nota:Remessa_JET}',
			'{observacao:Remessa_JET}',
			'{meta:_correios_tracking_code}',
			'{meta:chave_do_campo}',
		);
	}

	public static function product_variables() {
		return array(
			'{produto_id}',
			'{nome_produto}',
			'{preco_produto}',
			'{link_produto}',
			'{product_id}',
			'{product_name}',
			'{product_price}',
			'{product_url}',
			'{sku}',
			'{estoque}',
			'{stock_quantity}',
			'{nome_site}',
			'{url_site}',
			'{site_name}',
			'{site_url}',
		);
	}
}
