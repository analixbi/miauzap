<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_WooCommerce {
	public function __construct() {
		add_action('woocommerce_order_status_changed', array($this, 'order_status_changed'), 10, 4);
		add_action('transition_post_status', array($this, 'product_published'), 10, 3);
		add_action('woocommerce_product_set_stock_status', array($this, 'product_stock_status_changed'), 10, 3);
		add_action('woocommerce_after_order_notes', array($this, 'checkout_optin_field'));
		add_action('woocommerce_checkout_update_order_meta', array($this, 'save_checkout_optin'));
	}

	public function order_status_changed($order_id, $old_status, $new_status, $order) {
		if (!$order instanceof WC_Order) {
			$order = wc_get_order($order_id);
		}
		if (!$order) {
			return;
		}

		$status_key = 'wc-' . sanitize_key($new_status);
		$templates = MZW_Options::get_order_templates();
		if (empty($templates[$status_key]['enabled'])) {
			return;
		}

		$steps = self::order_template_messages($templates[$status_key]);
		if (empty($steps)) {
			return;
		}

		$phone = self::sanitize_phone_digits($order->get_billing_phone());
		if ('' === $phone) {
			return;
		}

		$modified = $order->get_date_modified() ? $order->get_date_modified()->getTimestamp() : time();
		foreach ($steps as $index => $step) {
			$message = self::render_order_template($step['message'], $order, $new_status);
			if ('' === trim($message)) {
				continue;
			}

			MZW_Queue::enqueue(
				'order_status',
				$phone,
				$message,
				$order->get_id(),
				array(
					'dedupe_key'  => sprintf('order:%d:%s:%d:%s', $order->get_id(), $status_key, $index, $modified),
					'extra_delay' => absint($step['wait'] ?? 0),
				)
			);
		}
	}

	public function product_published($new_status, $old_status, $post) {
		if ('publish' !== $new_status || 'publish' === $old_status || 'product' !== $post->post_type) {
			return;
		}

		$marketing = MZW_Options::get_marketing();
		if (empty($marketing['new_product_enabled']) || empty($marketing['new_product_message'])) {
			return;
		}

		$product = wc_get_product($post->ID);
		if (!$product) {
			return;
		}

		$message = self::render_template($marketing['new_product_message'], self::product_variables($product));
		$this->enqueue_marketing_to_recent_customers('new_product', $message, $product->get_id(), 'new_product:' . $product->get_id());
	}

	public function product_stock_status_changed($product_id, $stock_status, $product = null) {
		if ('instock' !== $stock_status) {
			return;
		}

		$marketing = MZW_Options::get_marketing();
		if (empty($marketing['stock_enabled']) || empty($marketing['stock_message'])) {
			return;
		}

		if (!$product instanceof WC_Product) {
			$product = wc_get_product($product_id);
		}
		if (!$product) {
			return;
		}

		$message = self::render_template($marketing['stock_message'], self::product_variables($product));
		$this->enqueue_marketing_to_recent_customers('stock_update', $message, $product->get_id(), 'stock:' . $product->get_id() . ':' . gmdate('Ymd'));
	}

	public function checkout_optin_field($checkout) {
		$settings = MZW_Options::get_settings();
		if (empty($settings['marketing_require_optin'])) {
			return;
		}

		woocommerce_form_field(
			'mzw_marketing_optin',
			array(
				'type'     => 'checkbox',
				'class'    => array('form-row-wide'),
				'label'    => __('Quero receber avisos e ofertas desta loja pelo WhatsApp.', 'miauzap-woocommerce'),
				'required' => false,
			),
			$checkout->get_value('mzw_marketing_optin')
		);
	}

	public function save_checkout_optin($order_id) {
		$order = wc_get_order($order_id);
		if (!$order) {
			return;
		}

		$optin = !empty($_POST['mzw_marketing_optin']) ? 'yes' : 'no';
		$order->update_meta_data('_mzw_marketing_optin', $optin);
		$order->save();

		$user_id = $order->get_user_id();
		if ($user_id) {
			update_user_meta($user_id, 'mzw_marketing_optin', $optin);
			self::save_user_phone_meta($user_id, $order->get_billing_phone());
		}
	}

	public function enqueue_marketing_to_recent_customers($type, $message, $related_id = 0, $dedupe_prefix = '') {
		foreach (self::recent_customer_phones() as $phone) {
			MZW_Queue::enqueue(
				$type,
				$phone,
				$message,
				$related_id,
				array('dedupe_key' => $dedupe_prefix . ':' . md5($phone))
			);
		}
	}

	public static function recent_customer_phones() {
		if (!function_exists('wc_get_orders')) {
			return array();
		}

		$settings = MZW_Options::get_settings();
		$orders = wc_get_orders(array(
			'limit'   => max(1, absint($settings['marketing_recipient_limit'])),
			'status'  => array('processing', 'completed'),
			'orderby' => 'date',
			'order'   => 'DESC',
			'return'  => 'objects',
		));

		$phones = array();
		foreach ($orders as $order) {
			if (!$order instanceof WC_Order) {
				continue;
			}

			if (!empty($settings['marketing_require_optin'])) {
				$user_optin = $order->get_user_id() ? get_user_meta($order->get_user_id(), 'mzw_marketing_optin', true) : '';
				$order_optin = $order->get_meta('_mzw_marketing_optin');
				if ('yes' !== $user_optin && 'yes' !== $order_optin) {
					continue;
				}
			}

			$phone = self::sanitize_phone_digits($order->get_billing_phone());
			if ($phone) {
				$phones[$phone] = $phone;
			}
		}

		return array_values($phones);
	}

	public static function render_order_template($template, $order, $status = '') {
		$template = self::replace_order_note_variables($template, $order);
		$template = self::replace_order_meta_variables($template, $order);
		return self::render_template($template, self::order_variables($order, $status));
	}

	public static function order_template_messages($template) {
		$messages = array();
		if (is_array($template) && !empty($template['messages']) && is_array($template['messages'])) {
			foreach ($template['messages'] as $step) {
				if (is_string($step)) {
					$step = array('message' => $step);
				}
				if (!is_array($step)) {
					continue;
				}

				$message = isset($step['message']) ? (string) $step['message'] : '';
				if ('' === trim($message)) {
					continue;
				}

				$messages[] = array(
					'message' => $message,
					'wait'    => absint($step['wait'] ?? 0),
				);
			}
		}

		if (empty($messages) && is_array($template) && !empty($template['message'])) {
			$messages[] = array(
				'message' => (string) $template['message'],
				'wait'    => 0,
			);
		}
		if (!empty($messages)) {
			$messages[0]['wait'] = 0;
		}

		return $messages;
	}

	public static function order_variables($order, $status = '') {
		$items = array();
		foreach ($order->get_items() as $item) {
			$items[] = $item->get_name() . ' x ' . $item->get_quantity();
		}

		$status_label = $status;
		if (function_exists('wc_get_order_status_name')) {
			$status_label = wc_get_order_status_name($status);
		}

		return array(
			'{pedido_id}'       => $order->get_id(),
			'{numero_pedido}'   => $order->get_order_number(),
			'{order_id}'        => $order->get_id(),
			'{order_number}'    => $order->get_order_number(),
			'{status}'          => $status_label,
			'{nome_cliente}'    => trim($order->get_billing_first_name() . ' ' . $order->get_billing_last_name()),
			'{primeiro_nome}'   => $order->get_billing_first_name(),
			'{sobrenome}'       => $order->get_billing_last_name(),
			'{customer_name}'   => trim($order->get_billing_first_name() . ' ' . $order->get_billing_last_name()),
			'{first_name}'      => $order->get_billing_first_name(),
			'{last_name}'       => $order->get_billing_last_name(),
			'{total}'           => self::format_price($order->get_total(), $order->get_currency()),
			'{moeda}'           => $order->get_currency(),
			'{pagamento}'       => $order->get_payment_method_title(),
			'{entrega}'         => $order->get_shipping_method(),
			'{itens}'           => implode(', ', $items),
			'{currency}'        => $order->get_currency(),
			'{payment_method}'  => $order->get_payment_method_title(),
			'{shipping_method}' => $order->get_shipping_method(),
			'{items}'           => implode(', ', $items),
			'{telefone}'        => $order->get_billing_phone(),
			'{email}'           => $order->get_billing_email(),
			'{data_pedido}'     => $order->get_date_created() ? $order->get_date_created()->date_i18n(get_option('date_format')) : '',
			'{link_pedido}'     => $order->get_view_order_url(),
			'{correios_tracking_code}' => self::order_meta_value($order, '_correios_tracking_code'),
			'{codigo_rastreio}' => self::order_meta_value($order, '_correios_tracking_code'),
			'{billing_phone}'   => $order->get_billing_phone(),
			'{billing_email}'   => $order->get_billing_email(),
			'{order_date}'      => $order->get_date_created() ? $order->get_date_created()->date_i18n(get_option('date_format')) : '',
			'{order_url}'       => $order->get_view_order_url(),
			'{nome_site}'       => get_bloginfo('name'),
			'{url_site}'        => home_url('/'),
			'{site_name}'       => get_bloginfo('name'),
			'{site_url}'        => home_url('/'),
		);
	}

	public static function product_variables($product) {
		return array(
			'{produto_id}'       => $product->get_id(),
			'{nome_produto}'     => $product->get_name(),
			'{preco_produto}'    => self::plain_text($product->get_price_html()),
			'{link_produto}'     => get_permalink($product->get_id()),
			'{product_id}'       => $product->get_id(),
			'{product_name}'     => $product->get_name(),
			'{product_price}'    => self::plain_text($product->get_price_html()),
			'{product_url}'      => get_permalink($product->get_id()),
			'{sku}'              => $product->get_sku(),
			'{estoque}'          => $product->get_stock_quantity(),
			'{stock_quantity}'   => $product->get_stock_quantity(),
			'{nome_site}'        => get_bloginfo('name'),
			'{url_site}'         => home_url('/'),
			'{site_name}'        => get_bloginfo('name'),
			'{site_url}'         => home_url('/'),
		);
	}

	public static function render_template($template, $variables) {
		$template = (string) $template;
		foreach ($variables as $key => $value) {
			$template = str_replace($key, (string) $value, $template);
		}

		return $template;
	}

	public static function plain_text($value) {
		$text = wp_strip_all_tags((string) $value);
		$charset = get_bloginfo('charset');
		$text = html_entity_decode($text, ENT_QUOTES | ENT_HTML5, $charset ? $charset : 'UTF-8');
		$text = str_replace(array('&nbsp;', '&#160;', '&#xA0;', "\xc2\xa0"), ' ', $text);
		$text = preg_replace('/[ \t]+/', ' ', $text);

		return trim($text);
	}

	public static function format_price($amount, $currency = '') {
		if (function_exists('wc_price')) {
			return self::plain_text(wc_price($amount, array('currency' => $currency)));
		}

		return self::plain_text((string) $amount);
	}

	public static function normalize_phone($phone) {
		$variants = self::phone_variants($phone);
		return empty($variants) ? '' : $variants[0];
	}

	public static function sanitize_phone_digits($phone) {
		$digits = preg_replace('/\D+/', '', (string) $phone);
		if ('' === $digits) {
			return '';
		}

		if (0 === strpos($digits, '00')) {
			$digits = substr($digits, 2);
		}

		return ltrim($digits, '0');
	}

	public static function phone_variants($phone) {
		$digits = self::sanitize_phone_digits($phone);
		if ('' === $digits) {
			return array();
		}

		$settings = MZW_Options::get_settings();
		$country = preg_replace('/\D+/', '', (string) $settings['default_country_code']);
		$with_country = $digits;

		if ($country && strlen($digits) <= 11 && 0 !== strpos($digits, $country)) {
			$with_country = $country . $digits;
		}

		$variants = array($with_country);
		if (0 === strpos($with_country, '55')) {
			$national = substr($with_country, 2);
			if (10 === strlen($national)) {
				$variants[] = '55' . substr($national, 0, 2) . '9' . substr($national, 2);
			} elseif (11 === strlen($national) && '9' === substr($national, 2, 1)) {
				$variants[] = '55' . substr($national, 0, 2) . substr($national, 3);
			}
		}

		if ($digits !== $with_country) {
			$variants[] = $digits;
		}

		return array_values(array_unique(array_filter($variants)));
	}

	public static function save_user_phone_meta($user_id, $phone) {
		$variants = self::phone_variants($phone);
		if (empty($variants)) {
			return;
		}

		update_user_meta($user_id, 'mzw_phone_e164', $variants[0]);
		update_user_meta($user_id, 'mzw_phone_variants', implode(',', $variants));

		foreach ($variants as $variant) {
			if (0 === strpos($variant, '55') && 13 === strlen($variant)) {
				update_user_meta($user_id, 'mzw_phone_com_9', $variant);
			}
			if (0 === strpos($variant, '55') && 12 === strlen($variant)) {
				update_user_meta($user_id, 'mzw_phone_sem_9', $variant);
			}
		}
	}

	public static function replace_order_note_variables($template, $order) {
		if (!$order || false === strpos((string) $template, '{')) {
			return $template;
		}

		return preg_replace_callback('/\{(nota|observacao|observacao_cliente|observacao_privada|nota_cliente|nota_privada|order_note|customer_note|private_note):([^}]+)\}/i', function ($matches) use ($order) {
			$type = strtolower($matches[1]);
			$keyword = trim($matches[2]);
			$visibility = 'any';

			if (in_array($type, array('nota_cliente', 'observacao_cliente', 'customer_note'), true)) {
				$visibility = 'customer';
			} elseif (in_array($type, array('nota_privada', 'observacao_privada', 'private_note'), true)) {
				$visibility = 'private';
			}

			return self::find_order_note_text($order, $keyword, $visibility);
		}, (string) $template);
	}

	public static function replace_order_meta_variables($template, $order) {
		if (!$order || false === strpos((string) $template, '{')) {
			return $template;
		}

		return preg_replace_callback('/\{(meta|pedido_meta|order_meta):([^}]+)\}/i', function ($matches) use ($order) {
			$key = trim($matches[2]);
			return self::order_meta_value($order, $key);
		}, (string) $template);
	}

	public static function order_meta_value($order, $key) {
		if (!$order || '' === trim((string) $key)) {
			return '';
		}

		$value = $order->get_meta(trim((string) $key), true);
		if (is_array($value)) {
			$value = implode(', ', array_map('strval', $value));
		} elseif (is_object($value)) {
			$value = wp_json_encode($value);
		}

		return self::plain_text((string) $value);
	}

	public static function find_order_note_text($order, $keyword, $visibility = 'any') {
		if (!function_exists('wc_get_order_notes') || !$order || '' === $keyword) {
			return '';
		}

		$notes = wc_get_order_notes(array(
			'order_id' => $order->get_id(),
			'orderby' => 'date_created',
			'order'   => 'DESC',
		));

		foreach ($notes as $note) {
			$content = '';
			$is_customer = false;

			if (is_object($note) && method_exists($note, 'get_content')) {
				$content = $note->get_content();
				$is_customer = method_exists($note, 'get_customer_note') ? (bool) $note->get_customer_note() : false;
			} elseif (is_object($note)) {
				$content = isset($note->content) ? $note->content : '';
				$is_customer = !empty($note->customer_note);
			}

			if ('customer' === $visibility && !$is_customer) {
				continue;
			}
			if ('private' === $visibility && $is_customer) {
				continue;
			}

			$text = trim(wp_strip_all_tags((string) $content));
			if ('' === $text || false === stripos($text, $keyword)) {
				continue;
			}

			return self::extract_note_keyword_text($text, $keyword);
		}

		return '';
	}

	private static function extract_note_keyword_text($text, $keyword) {
		$quoted = preg_quote($keyword, '/');
		foreach (preg_split('/\r\n|\r|\n/', $text) as $line) {
			if (preg_match('/' . $quoted . '\s*[:=\-]\s*(.+)$/iu', $line, $match)) {
				return trim($match[1]);
			}
		}

		return $text;
	}

	public static function latest_order() {
		if (!function_exists('wc_get_orders')) {
			return null;
		}

		$orders = wc_get_orders(array(
			'limit'   => 1,
			'orderby' => 'date',
			'order'   => 'DESC',
			'return'  => 'objects',
		));

		return !empty($orders) ? $orders[0] : null;
	}
}
