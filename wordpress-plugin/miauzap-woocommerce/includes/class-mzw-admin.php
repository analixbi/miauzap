<?php

if (!defined('ABSPATH')) {
	exit;
}

class MZW_Admin {
	private $page_slug = 'miauzap-woocommerce';

	public function __construct() {
		add_action('admin_menu', array($this, 'menu'));
		add_action('admin_enqueue_scripts', array($this, 'enqueue_assets'));
		add_action('wp_ajax_mzw_preview_order_template', array($this, 'ajax_preview_order_template'));
		add_action('wp_ajax_mzw_test_instance', array($this, 'ajax_test_instance'));
	}

	public function menu() {
		add_menu_page(
			__('Miauzap', 'miauzap-woocommerce'),
			__('Miauzap', 'miauzap-woocommerce'),
			'manage_woocommerce',
			$this->page_slug,
			array($this, 'render'),
			'dashicons-email-alt',
			56
		);
	}

	public function enqueue_assets($hook) {
		if ('toplevel_page_' . $this->page_slug !== $hook) {
			return;
		}

		wp_enqueue_style('mzw-admin', MZW_URL . 'assets/admin.css', array(), MZW_VERSION);
		wp_enqueue_script('mzw-admin', MZW_URL . 'assets/admin.js', array('jquery'), MZW_VERSION, true);
		wp_localize_script(
			'mzw-admin',
			'MZWAdmin',
			array(
				'ajaxUrl' => admin_url('admin-ajax.php'),
				'nonce'   => wp_create_nonce('mzw_admin'),
				'vars'    => MZW_Options::order_variables(),
			)
		);
	}

	public function render() {
		if (!current_user_can('manage_woocommerce')) {
			wp_die(esc_html__('Voce nao tem permissao para gerenciar o Miauzap.', 'miauzap-woocommerce'));
		}

		$this->handle_post();

		$tab = sanitize_key($_GET['tab'] ?? 'connection');
		$tabs = array(
			'connection' => __('Conexao', 'miauzap-woocommerce'),
			'orders'     => __('Mensagens de pedidos', 'miauzap-woocommerce'),
			'marketing'  => __('Marketing', 'miauzap-woocommerce'),
			'queue'      => __('Fila', 'miauzap-woocommerce'),
		);
		if (!isset($tabs[$tab])) {
			$tab = 'connection';
		}

		echo '<div class="wrap mzw-wrap">';
		echo '<h1>' . esc_html__('Miauzap para WooCommerce', 'miauzap-woocommerce') . '</h1>';
		settings_errors('mzw');

		echo '<nav class="nav-tab-wrapper">';
		foreach ($tabs as $key => $label) {
			$url = admin_url('admin.php?page=' . $this->page_slug . '&tab=' . $key);
			$class = $tab === $key ? ' nav-tab-active' : '';
			echo '<a class="nav-tab' . esc_attr($class) . '" href="' . esc_url($url) . '">' . esc_html($label) . '</a>';
		}
		echo '</nav>';

		if ('orders' === $tab) {
			$this->render_orders();
		} elseif ('marketing' === $tab) {
			$this->render_marketing();
		} elseif ('queue' === $tab) {
			$this->render_queue();
		} else {
			$this->render_connection();
		}

		echo '</div>';
	}

	private function handle_post() {
		if ('POST' !== strtoupper($_SERVER['REQUEST_METHOD'] ?? '') || empty($_POST['mzw_action'])) {
			return;
		}

		check_admin_referer('mzw_save');
		$action = sanitize_key($_POST['mzw_action']);

		if ('save_connection' === $action) {
			$this->save_connection();
		} elseif ('save_orders' === $action) {
			$this->save_orders();
		} elseif ('save_marketing' === $action) {
			$this->save_marketing();
		} elseif ('create_campaign' === $action) {
			$this->create_campaign();
		} elseif ('process_queue' === $action) {
			MZW_Queue::process_due(1);
			add_settings_error('mzw', 'queue_processed', __('Fila acionada. Os proximos itens seguem o intervalo configurado.', 'miauzap-woocommerce'), 'updated');
		} elseif ('retry_queue' === $action && !empty($_POST['queue_id'])) {
			MZW_Queue::retry(absint($_POST['queue_id']));
			add_settings_error('mzw', 'queue_retried', __('Item retornou para a fila.', 'miauzap-woocommerce'), 'updated');
		}
	}

	private function save_connection() {
		$current = MZW_Options::get_settings();
		$settings = $current;

		$text_fields = array('base_url', 'default_country_code', 'quiet_hours_start', 'quiet_hours_end', 'otp_template');
		foreach ($text_fields as $field) {
			if (isset($_POST[$field])) {
				$settings[$field] = 'otp_template' === $field ? sanitize_textarea_field(wp_unslash($_POST[$field])) : sanitize_text_field(wp_unslash($_POST[$field]));
			}
		}

		$int_fields = array('min_delay', 'max_delay', 'otp_min_delay', 'otp_max_delay', 'process_limit', 'max_attempts', 'daily_limit_per_instance', 'marketing_recipient_limit', 'otp_length', 'otp_ttl_minutes');
		foreach ($int_fields as $field) {
			if (isset($_POST[$field])) {
				$settings[$field] = absint($_POST[$field]);
			}
		}

		foreach (array('check_whatsapp', 'link_preview', 'quiet_hours_enabled', 'marketing_require_optin', 'otp_enabled') as $field) {
			$settings[$field] = !empty($_POST[$field]) ? 1 : 0;
		}
		$settings['otp_create_customer'] = 0;

		$settings['instances'] = $this->sanitize_instances($_POST['instances'] ?? array(), $current['instances'] ?? array());
		MZW_Options::update_settings($settings);
		add_settings_error('mzw', 'settings_saved', __('Configuracoes de conexao salvas.', 'miauzap-woocommerce'), 'updated');
	}

	private function sanitize_instances($posted, $existing) {
		$out = array();
		if (!is_array($posted)) {
			return $out;
		}

		foreach ($posted as $index => $row) {
			if (!is_array($row)) {
				continue;
			}

			$label = sanitize_text_field(wp_unslash($row['label'] ?? ''));
			$key = sanitize_key(wp_unslash($row['key'] ?? ''));
			if (!$key) {
				$key = sanitize_key($label ? $label : 'instance_' . $index);
			}
			if (isset($out[$key])) {
				$key .= '_' . absint($index);
			}

			$token = sanitize_text_field(wp_unslash($row['token'] ?? ''));
			if ('' === $token && isset($existing[$key]['token'])) {
				$token = $existing[$key]['token'];
			}
			if ('' === $token && '' === $label) {
				continue;
			}

			$out[$key] = array(
				'label'   => $label ? $label : $key,
				'token'   => $token,
				'weight'  => max(1, absint($row['weight'] ?? 1)),
				'enabled' => !empty($row['enabled']) ? 1 : 0,
			);
		}

		return $out;
	}

	private function save_orders() {
		$templates = array();
		$posted = $_POST['order_templates'] ?? array();
		if (is_array($posted)) {
			foreach ($posted as $status => $row) {
				$status = sanitize_key($status);
				$messages = array();
				if (!empty($row['messages']) && is_array($row['messages'])) {
					foreach ($row['messages'] as $step) {
						if (!is_array($step)) {
							continue;
						}

						$raw_message = $step['message'] ?? '';
						if (is_array($raw_message)) {
							continue;
						}

						$message = sanitize_textarea_field(wp_unslash($raw_message));
						if ('' === trim($message)) {
							continue;
						}

						$messages[] = array(
							'message' => $message,
							'wait'    => absint($step['wait'] ?? 0),
						);
					}
				}

				if (empty($messages)) {
					$legacy_message = sanitize_textarea_field(wp_unslash($row['message'] ?? ''));
					if ('' !== trim($legacy_message)) {
						$messages[] = array(
							'message' => $legacy_message,
							'wait'    => 0,
						);
					}
				}
				if (!empty($messages)) {
					$messages[0]['wait'] = 0;
				}

				$templates[$status] = array(
					'enabled'  => !empty($row['enabled']) ? 1 : 0,
					'message'  => isset($messages[0]['message']) ? $messages[0]['message'] : '',
					'messages' => $messages,
				);
			}
		}

		MZW_Options::update_order_templates($templates);
		add_settings_error('mzw', 'templates_saved', __('Modelos de mensagens de pedido salvos.', 'miauzap-woocommerce'), 'updated');
	}

	private function save_marketing() {
		$marketing = MZW_Options::get_marketing();
		$marketing['new_product_enabled'] = !empty($_POST['new_product_enabled']) ? 1 : 0;
		$marketing['stock_enabled'] = !empty($_POST['stock_enabled']) ? 1 : 0;
		$marketing['new_product_message'] = sanitize_textarea_field(wp_unslash($_POST['new_product_message'] ?? ''));
		$marketing['stock_message'] = sanitize_textarea_field(wp_unslash($_POST['stock_message'] ?? ''));

		MZW_Options::update_marketing($marketing);
		add_settings_error('mzw', 'marketing_saved', __('Configuracoes de marketing salvas.', 'miauzap-woocommerce'), 'updated');
	}

	private function create_campaign() {
		$message = sanitize_textarea_field(wp_unslash($_POST['campaign_message'] ?? ''));
		$product_id = absint($_POST['campaign_product_id'] ?? 0);
		$phones_text = sanitize_textarea_field(wp_unslash($_POST['campaign_phones'] ?? ''));
		$use_recent = !empty($_POST['campaign_recent_customers']);

		if ($product_id && function_exists('wc_get_product')) {
			$product = wc_get_product($product_id);
			if ($product) {
				$message = MZW_WooCommerce::render_template($message, MZW_WooCommerce::product_variables($product));
			}
		}

		$phones = array();
		if ($phones_text) {
			foreach (preg_split('/[\r\n,;]+/', $phones_text) as $phone) {
				$normalized = MZW_WooCommerce::sanitize_phone_digits($phone);
				if ($normalized) {
					$phones[$normalized] = $normalized;
				}
			}
		}
		if ($use_recent) {
			foreach (MZW_WooCommerce::recent_customer_phones() as $phone) {
				$phones[$phone] = $phone;
			}
		}

		if (!$message || empty($phones)) {
			add_settings_error('mzw', 'campaign_empty', __('A campanha precisa de uma mensagem e pelo menos um telefone.', 'miauzap-woocommerce'), 'error');
			return;
		}

		$count = 0;
		foreach ($phones as $phone) {
			$result = MZW_Queue::enqueue('manual_campaign', $phone, $message, $product_id, array('dedupe_key' => 'campaign:' . md5($message . '|' . $phone . '|' . current_time('mysql'))));
			if (!is_wp_error($result)) {
				$count++;
			}
		}

		add_settings_error('mzw', 'campaign_created', sprintf(__('%d mensagens adicionadas a fila.', 'miauzap-woocommerce'), $count), 'updated');
	}

	private function render_connection() {
		$settings = MZW_Options::get_settings();
		$instances = MZW_Options::get_instances(false);
		?>
		<form method="post" class="mzw-panel">
			<?php wp_nonce_field('mzw_save'); ?>
			<input type="hidden" name="mzw_action" value="save_connection">

			<h2><?php echo esc_html__('Conexao com a API', 'miauzap-woocommerce'); ?></h2>
			<table class="form-table" role="presentation">
				<tr>
					<th scope="row"><label for="base_url"><?php echo esc_html__('URL base do Miauzap', 'miauzap-woocommerce'); ?></label></th>
					<td><input type="url" class="regular-text" id="base_url" name="base_url" value="<?php echo esc_attr($settings['base_url']); ?>" placeholder="https://api.exemplo.com"></td>
				</tr>
				<tr>
					<th scope="row"><label for="default_country_code"><?php echo esc_html__('Codigo do pais padrao', 'miauzap-woocommerce'); ?></label></th>
					<td><input type="text" id="default_country_code" name="default_country_code" value="<?php echo esc_attr($settings['default_country_code']); ?>" class="small-text"></td>
				</tr>
			</table>

			<h2><?php echo esc_html__('Instancias', 'miauzap-woocommerce'); ?></h2>
			<table class="widefat striped mzw-instances">
				<thead>
					<tr>
						<th><?php echo esc_html__('Ativa', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Nome', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Token', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Peso', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Acoes', 'miauzap-woocommerce'); ?></th>
					</tr>
				</thead>
				<tbody>
					<?php
					$rows = empty($instances) ? array('__new__' => array('label' => '', 'token' => '', 'weight' => 1, 'enabled' => 1, 'key' => '')) : $instances;
					$i = 0;
					foreach ($rows as $key => $instance) :
						?>
						<tr>
							<td><input type="checkbox" name="instances[<?php echo esc_attr($i); ?>][enabled]" value="1" <?php checked(!empty($instance['enabled'])); ?>></td>
							<td>
								<input type="hidden" name="instances[<?php echo esc_attr($i); ?>][key]" value="<?php echo esc_attr('__new__' === $key ? '' : $key); ?>">
								<input type="text" name="instances[<?php echo esc_attr($i); ?>][label]" value="<?php echo esc_attr($instance['label'] ?? ''); ?>" placeholder="principal">
							</td>
							<td><input type="password" name="instances[<?php echo esc_attr($i); ?>][token]" value="" placeholder="<?php echo esc_attr(!empty($instance['token']) ? __('salvo - deixe vazio para manter', 'miauzap-woocommerce') : __('cole o token', 'miauzap-woocommerce')); ?>" autocomplete="new-password"></td>
							<td><input type="number" min="1" name="instances[<?php echo esc_attr($i); ?>][weight]" value="<?php echo esc_attr($instance['weight'] ?? 1); ?>" class="small-text"></td>
							<td>
								<button type="button" class="button mzw-test-instance"><?php echo esc_html__('Testar', 'miauzap-woocommerce'); ?></button>
								<button type="button" class="button mzw-remove-row"><?php echo esc_html__('Remover', 'miauzap-woocommerce'); ?></button>
							</td>
						</tr>
						<?php
						$i++;
					endforeach;
					?>
				</tbody>
			</table>
			<p><button type="button" class="button" id="mzw-add-instance"><?php echo esc_html__('Adicionar instancia', 'miauzap-woocommerce'); ?></button></p>

			<h2><?php echo esc_html__('Controles para reduzir risco de bloqueio', 'miauzap-woocommerce'); ?></h2>
			<table class="form-table" role="presentation">
				<tr>
					<th scope="row"><?php echo esc_html__('Atraso aleatorio', 'miauzap-woocommerce'); ?></th>
					<td>
						<input type="number" min="0" name="min_delay" value="<?php echo esc_attr($settings['min_delay']); ?>" class="small-text"> -
						<input type="number" min="0" name="max_delay" value="<?php echo esc_attr($settings['max_delay']); ?>" class="small-text">
						<?php echo esc_html__('segundos entre mensagens da fila', 'miauzap-woocommerce'); ?>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Atraso do OTP', 'miauzap-woocommerce'); ?></th>
					<td>
						<input type="number" min="0" name="otp_min_delay" value="<?php echo esc_attr($settings['otp_min_delay']); ?>" class="small-text"> -
						<input type="number" min="0" name="otp_max_delay" value="<?php echo esc_attr($settings['otp_max_delay']); ?>" class="small-text">
						<?php echo esc_html__('segundos', 'miauzap-woocommerce'); ?>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Limites', 'miauzap-woocommerce'); ?></th>
					<td>
						<input type="number" min="1" name="process_limit" value="<?php echo esc_attr($settings['process_limit']); ?>" class="small-text"> <?php echo esc_html__('por execucao', 'miauzap-woocommerce'); ?>
						<input type="number" min="0" name="daily_limit_per_instance" value="<?php echo esc_attr($settings['daily_limit_per_instance']); ?>" class="small-text"> <?php echo esc_html__('por instancia/dia', 'miauzap-woocommerce'); ?>
						<input type="number" min="1" name="max_attempts" value="<?php echo esc_attr($settings['max_attempts']); ?>" class="small-text"> <?php echo esc_html__('tentativas', 'miauzap-woocommerce'); ?>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Opcoes', 'miauzap-woocommerce'); ?></th>
					<td>
						<label><input type="checkbox" name="check_whatsapp" value="1" <?php checked($settings['check_whatsapp']); ?>> <?php echo esc_html__('Validar numero no WhatsApp antes de enviar', 'miauzap-woocommerce'); ?></label><br>
						<label><input type="checkbox" name="link_preview" value="1" <?php checked($settings['link_preview']); ?>> <?php echo esc_html__('Ativar preview de links nas mensagens', 'miauzap-woocommerce'); ?></label><br>
						<label><input type="checkbox" name="marketing_require_optin" value="1" <?php checked($settings['marketing_require_optin']); ?>> <?php echo esc_html__('Exigir opt-in para marketing', 'miauzap-woocommerce'); ?></label>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Horario silencioso', 'miauzap-woocommerce'); ?></th>
					<td>
						<label><input type="checkbox" name="quiet_hours_enabled" value="1" <?php checked($settings['quiet_hours_enabled']); ?>> <?php echo esc_html__('Ativo', 'miauzap-woocommerce'); ?></label>
						<input type="time" name="quiet_hours_start" value="<?php echo esc_attr($settings['quiet_hours_start']); ?>">
						<input type="time" name="quiet_hours_end" value="<?php echo esc_attr($settings['quiet_hours_end']); ?>">
					</td>
				</tr>
			</table>

			<h2><?php echo esc_html__('Login por WhatsApp', 'miauzap-woocommerce'); ?></h2>
			<table class="form-table" role="presentation">
				<tr>
					<th scope="row"><?php echo esc_html__('Ativar', 'miauzap-woocommerce'); ?></th>
					<td>
						<label><input type="checkbox" name="otp_enabled" value="1" <?php checked($settings['otp_enabled']); ?>> <?php echo esc_html__('Permitir login com codigo enviado pelo WhatsApp', 'miauzap-woocommerce'); ?></label><br>
						<span class="description"><?php echo esc_html__('Somente clientes com telefone cadastrado em Telefone de cobranca conseguem entrar.', 'miauzap-woocommerce'); ?></span>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Codigo', 'miauzap-woocommerce'); ?></th>
					<td>
						<input type="number" min="4" max="8" name="otp_length" value="<?php echo esc_attr($settings['otp_length']); ?>" class="small-text"> <?php echo esc_html__('digitos, valido por', 'miauzap-woocommerce'); ?>
						<input type="number" min="1" name="otp_ttl_minutes" value="<?php echo esc_attr($settings['otp_ttl_minutes']); ?>" class="small-text"> <?php echo esc_html__('minutos', 'miauzap-woocommerce'); ?>
					</td>
				</tr>
				<tr>
					<th scope="row"><label for="otp_template"><?php echo esc_html__('Mensagem do OTP', 'miauzap-woocommerce'); ?></label></th>
					<td><textarea id="otp_template" name="otp_template" rows="4" class="large-text code"><?php echo esc_textarea($settings['otp_template']); ?></textarea></td>
				</tr>
			</table>

			<?php submit_button(__('Salvar configuracoes', 'miauzap-woocommerce')); ?>
		</form>
		<?php
	}

	private function render_orders() {
		$templates = MZW_Options::get_order_templates();
		$statuses = function_exists('wc_get_order_statuses') ? wc_get_order_statuses() : array();
		if (empty($statuses)) {
			echo '<div class="mzw-panel"><p>' . esc_html__('Os status de pedido do WooCommerce nao foram encontrados.', 'miauzap-woocommerce') . '</p></div>';
			return;
		}
		?>
		<form method="post" class="mzw-panel">
			<?php wp_nonce_field('mzw_save'); ?>
			<input type="hidden" name="mzw_action" value="save_orders">

			<h2><?php echo esc_html__('Mensagens por status do pedido', 'miauzap-woocommerce'); ?></h2>
			<p><?php echo esc_html__('Para puxar texto de anotacoes do pedido, use {nota:palavra}, {nota_cliente:palavra} ou {nota_privada:palavra}. Para campos personalizados do pedido, use {meta:nome_do_campo}. Exemplo: {correios_tracking_code} ou {meta:_correios_tracking_code}.', 'miauzap-woocommerce'); ?></p>
			<div class="mzw-orders-layout">
				<div class="mzw-orders-main">
					<?php $this->render_variable_chips(MZW_Options::order_variables()); ?>
					<div class="mzw-note-builder">
						<label for="mzw-note-keyword"><?php echo esc_html__('Criar variavel de anotacao', 'miauzap-woocommerce'); ?></label>
						<select id="mzw-note-scope">
							<option value="nota"><?php echo esc_html__('Qualquer anotacao', 'miauzap-woocommerce'); ?></option>
							<option value="nota_privada"><?php echo esc_html__('Somente anotacao privada', 'miauzap-woocommerce'); ?></option>
							<option value="nota_cliente"><?php echo esc_html__('Somente anotacao do cliente', 'miauzap-woocommerce'); ?></option>
						</select>
						<input type="text" id="mzw-note-keyword" placeholder="Remessa_JET">
						<button type="button" class="button" id="mzw-insert-note-var"><?php echo esc_html__('Inserir', 'miauzap-woocommerce'); ?></button>
					</div>

					<div class="mzw-status-grid">
						<?php foreach ($statuses as $status => $label) :
							$template = isset($templates[$status]) && is_array($templates[$status]) ? $templates[$status] : array();
							$steps = MZW_WooCommerce::order_template_messages($template);
							if (empty($steps)) {
								$steps = array(
									array(
										'message' => $this->default_order_template(),
										'wait'    => 0,
									),
								);
							}
							$enabled = !empty($templates[$status]['enabled']);
							?>
							<section class="mzw-card">
								<label class="mzw-card-title">
									<input type="checkbox" name="order_templates[<?php echo esc_attr($status); ?>][enabled]" value="1" <?php checked($enabled); ?>>
									<?php echo esc_html($label); ?>
									<code><?php echo esc_html($status); ?></code>
								</label>
								<div class="mzw-flow" data-status="<?php echo esc_attr($status); ?>">
									<?php foreach ($steps as $index => $step) : ?>
										<div class="mzw-flow-step">
											<div class="mzw-flow-step-head">
												<strong class="mzw-flow-step-title"><?php echo esc_html(sprintf(__('Mensagem %d', 'miauzap-woocommerce'), $index + 1)); ?></strong>
												<button type="button" class="button-link-delete mzw-remove-message" <?php echo 0 === $index ? 'hidden' : ''; ?>><?php echo esc_html__('Remover', 'miauzap-woocommerce'); ?></button>
											</div>
											<label class="mzw-flow-wait <?php echo 0 === $index ? 'mzw-flow-wait-hidden' : ''; ?>">
												<?php echo esc_html__('Aguardar antes desta mensagem', 'miauzap-woocommerce'); ?>
												<input type="number" min="0" step="1" class="small-text mzw-flow-wait-input" name="order_templates[<?php echo esc_attr($status); ?>][messages][<?php echo esc_attr($index); ?>][wait]" value="<?php echo esc_attr(absint($step['wait'] ?? 0)); ?>">
												<?php echo esc_html__('segundos', 'miauzap-woocommerce'); ?>
											</label>
											<textarea class="large-text code mzw-template" rows="7" name="order_templates[<?php echo esc_attr($status); ?>][messages][<?php echo esc_attr($index); ?>][message]"><?php echo esc_textarea($step['message']); ?></textarea>
										</div>
									<?php endforeach; ?>
								</div>
								<div class="mzw-flow-actions">
									<button type="button" class="button mzw-add-message"><?php echo esc_html__('Adicionar mensagem', 'miauzap-woocommerce'); ?></button>
									<button type="button" class="button mzw-preview-template" data-status="<?php echo esc_attr($status); ?>" data-status-label="<?php echo esc_attr($label); ?>"><?php echo esc_html__('Previsualizar', 'miauzap-woocommerce'); ?></button>
								</div>
							</section>
						<?php endforeach; ?>
					</div>
				</div>

				<aside class="mzw-preview-panel">
					<h2><?php echo esc_html__('Previsualizacao WhatsApp', 'miauzap-woocommerce'); ?></h2>
					<label class="mzw-preview-order">
						<?php echo esc_html__('ID do pedido', 'miauzap-woocommerce'); ?>
						<input type="number" id="mzw-preview-order-id" class="small-text" placeholder="<?php echo esc_attr__('mais recente', 'miauzap-woocommerce'); ?>">
					</label>
					<div class="mzw-phone-preview" aria-live="polite">
						<div class="mzw-phone-header">
							<div class="mzw-phone-avatar">M</div>
							<div>
								<strong><?php echo esc_html(get_bloginfo('name') ? get_bloginfo('name') : __('Loja', 'miauzap-woocommerce')); ?></strong>
								<span id="mzw-preview-status"><?php echo esc_html__('Selecione um status', 'miauzap-woocommerce'); ?></span>
							</div>
						</div>
						<div class="mzw-phone-chat" id="mzw-preview-output">
							<div class="mzw-whatsapp-bubble"><?php echo esc_html__('Clique em Previsualizar para carregar a mensagem aqui.', 'miauzap-woocommerce'); ?></div>
						</div>
					</div>
				</aside>
			</div>

			<?php submit_button(__('Salvar modelos', 'miauzap-woocommerce')); ?>
		</form>
		<?php
	}

	private function render_marketing() {
		$marketing = MZW_Options::get_marketing();
		?>
		<form method="post" class="mzw-panel">
			<?php wp_nonce_field('mzw_save'); ?>
			<input type="hidden" name="mzw_action" value="save_marketing">
			<h2><?php echo esc_html__('Mensagens automaticas de marketing', 'miauzap-woocommerce'); ?></h2>
			<?php $this->render_variable_chips(MZW_Options::product_variables()); ?>
			<table class="form-table" role="presentation">
				<tr>
					<th scope="row"><?php echo esc_html__('Produto novo', 'miauzap-woocommerce'); ?></th>
					<td>
						<label><input type="checkbox" name="new_product_enabled" value="1" <?php checked($marketing['new_product_enabled']); ?>> <?php echo esc_html__('Enviar quando um produto for publicado', 'miauzap-woocommerce'); ?></label>
						<textarea name="new_product_message" rows="5" class="large-text code mzw-template"><?php echo esc_textarea($marketing['new_product_message']); ?></textarea>
					</td>
				</tr>
				<tr>
					<th scope="row"><?php echo esc_html__('Atualizacao de estoque', 'miauzap-woocommerce'); ?></th>
					<td>
						<label><input type="checkbox" name="stock_enabled" value="1" <?php checked($marketing['stock_enabled']); ?>> <?php echo esc_html__('Enviar quando um produto voltar ao estoque', 'miauzap-woocommerce'); ?></label>
						<textarea name="stock_message" rows="5" class="large-text code mzw-template"><?php echo esc_textarea($marketing['stock_message']); ?></textarea>
					</td>
				</tr>
			</table>
			<?php submit_button(__('Salvar marketing', 'miauzap-woocommerce')); ?>
		</form>

		<form method="post" class="mzw-panel">
			<?php wp_nonce_field('mzw_save'); ?>
			<input type="hidden" name="mzw_action" value="create_campaign">
			<h2><?php echo esc_html__('Campanha manual', 'miauzap-woocommerce'); ?></h2>
			<p><?php echo esc_html__('Use para campanhas pequenas e segmentadas. As mensagens entram na fila randomizada e podem alternar entre instancias.', 'miauzap-woocommerce'); ?></p>
			<table class="form-table" role="presentation">
				<tr>
					<th scope="row"><label for="campaign_product_id"><?php echo esc_html__('ID do produto para variaveis', 'miauzap-woocommerce'); ?></label></th>
					<td><input type="number" min="0" id="campaign_product_id" name="campaign_product_id" class="small-text"></td>
				</tr>
				<tr>
					<th scope="row"><label for="campaign_message"><?php echo esc_html__('Mensagem', 'miauzap-woocommerce'); ?></label></th>
					<td><textarea id="campaign_message" name="campaign_message" rows="6" class="large-text code mzw-template"></textarea></td>
				</tr>
				<tr>
					<th scope="row"><label for="campaign_phones"><?php echo esc_html__('Telefones', 'miauzap-woocommerce'); ?></label></th>
					<td>
						<textarea id="campaign_phones" name="campaign_phones" rows="6" class="large-text" placeholder="5511999999999&#10;5511888888888"></textarea>
						<p><label><input type="checkbox" name="campaign_recent_customers" value="1"> <?php echo esc_html__('Tambem enviar para clientes recentes com opt-in', 'miauzap-woocommerce'); ?></label></p>
					</td>
				</tr>
			</table>
			<?php submit_button(__('Adicionar campanha a fila', 'miauzap-woocommerce'), 'primary'); ?>
		</form>
		<?php
	}

	private function render_queue() {
		$stats = MZW_Queue::stats();
		$status = sanitize_key($_GET['status'] ?? '');
		$items = MZW_Queue::get_items($status, 80);
		?>
		<div class="mzw-panel">
			<h2><?php echo esc_html__('Fila', 'miauzap-woocommerce'); ?></h2>
			<div class="mzw-stats">
				<?php foreach ($stats as $key => $qty) : ?>
					<a class="mzw-stat" href="<?php echo esc_url(admin_url('admin.php?page=' . $this->page_slug . '&tab=queue&status=' . $key)); ?>">
						<strong><?php echo esc_html($qty); ?></strong>
						<span><?php echo esc_html($this->status_label($key)); ?></span>
					</a>
				<?php endforeach; ?>
			</div>
			<form method="post">
				<?php wp_nonce_field('mzw_save'); ?>
				<input type="hidden" name="mzw_action" value="process_queue">
				<?php submit_button(__('Acordar fila agora', 'miauzap-woocommerce'), 'secondary', 'submit', false); ?>
			</form>
			<table class="widefat striped">
				<thead>
					<tr>
						<th>ID</th>
						<th><?php echo esc_html__('Status', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Tipo', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Telefone', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Instancia', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Agendada para', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Tentativas', 'miauzap-woocommerce'); ?></th>
						<th><?php echo esc_html__('Ultimo erro', 'miauzap-woocommerce'); ?></th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					<?php foreach ($items as $item) : ?>
						<tr>
							<td><?php echo esc_html($item['id']); ?></td>
							<td><?php echo esc_html($this->status_label($item['status'])); ?></td>
							<td><?php echo esc_html($this->type_label($item['type'])); ?></td>
							<td><?php echo esc_html($item['phone']); ?></td>
							<td><?php echo esc_html($this->instance_label($item['instance_key'])); ?></td>
							<td><?php echo esc_html(get_date_from_gmt($item['scheduled_at'])); ?></td>
							<td><?php echo esc_html($item['attempts']); ?></td>
							<td><?php echo esc_html(wp_trim_words($item['last_error'], 12)); ?></td>
							<td>
								<?php if (in_array($item['status'], array('failed', 'skipped'), true)) : ?>
									<form method="post">
										<?php wp_nonce_field('mzw_save'); ?>
										<input type="hidden" name="mzw_action" value="retry_queue">
										<input type="hidden" name="queue_id" value="<?php echo esc_attr($item['id']); ?>">
										<button class="button button-small"><?php echo esc_html__('Tentar novamente', 'miauzap-woocommerce'); ?></button>
									</form>
								<?php endif; ?>
							</td>
						</tr>
					<?php endforeach; ?>
				</tbody>
			</table>
		</div>
		<?php
	}

	private function render_variable_chips($variables) {
		echo '<div class="mzw-vars" aria-label="' . esc_attr__('Variaveis', 'miauzap-woocommerce') . '">';
		foreach ($variables as $var) {
			echo '<button type="button" class="button mzw-var" draggable="true" data-var="' . esc_attr($var) . '">' . esc_html($var) . '</button>';
		}
		echo '</div>';
	}

	private function default_order_template() {
		return "Ola {primeiro_nome}, seu pedido {numero_pedido} agora esta com status: {status}.\nTotal: {total}\n{nome_site}";
	}

	private function status_label($status) {
		$labels = array(
			'pending'    => __('Pendente', 'miauzap-woocommerce'),
			'processing' => __('Processando', 'miauzap-woocommerce'),
			'sent'       => __('Enviada', 'miauzap-woocommerce'),
			'failed'     => __('Falhou', 'miauzap-woocommerce'),
			'skipped'    => __('Ignorada', 'miauzap-woocommerce'),
		);

		return $labels[$status] ?? $status;
	}

	private function type_label($type) {
		$labels = array(
			'order_status'    => __('Status de pedido', 'miauzap-woocommerce'),
			'new_product'     => __('Produto novo', 'miauzap-woocommerce'),
			'stock_update'    => __('Estoque', 'miauzap-woocommerce'),
			'manual_campaign' => __('Campanha manual', 'miauzap-woocommerce'),
			'otp'             => __('OTP', 'miauzap-woocommerce'),
		);

		return $labels[$type] ?? $type;
	}

	private function instance_label($key) {
		$key = sanitize_key($key);
		if ('' === $key) {
			return '-';
		}

		$instance = MZW_Options::get_instance($key);
		if ($instance && !empty($instance['label'])) {
			return $instance['label'];
		}

		return $key;
	}

	public function ajax_preview_order_template() {
		check_ajax_referer('mzw_admin', 'nonce');
		if (!current_user_can('manage_woocommerce')) {
			wp_send_json_error(array('message' => __('Permissao negada.', 'miauzap-woocommerce')), 403);
		}
		if (!function_exists('wc_get_order')) {
			wp_send_json_error(array('message' => __('WooCommerce nao esta disponivel.', 'miauzap-woocommerce')), 500);
		}

		$status = sanitize_key($_POST['status'] ?? 'wc-pending');
		$order_id = absint($_POST['order_id'] ?? 0);
		$order = $order_id ? wc_get_order($order_id) : MZW_WooCommerce::latest_order();

		if (!$order) {
			wp_send_json_error(array('message' => __('Nenhum pedido encontrado para previsualizar.', 'miauzap-woocommerce')), 404);
		}

		$messages = array();
		if (!empty($_POST['messages']) && is_array($_POST['messages'])) {
			$posted_messages = wp_unslash($_POST['messages']);
			$posted_waits = isset($_POST['waits']) && is_array($_POST['waits']) ? wp_unslash($_POST['waits']) : array();
			foreach ($posted_messages as $index => $message) {
				if (is_array($message)) {
					continue;
				}

				$message = sanitize_textarea_field($message);
				if ('' === trim($message)) {
					continue;
				}

				$messages[] = array(
					'message' => MZW_WooCommerce::render_order_template($message, $order, str_replace('wc-', '', $status)),
					'wait'    => absint($posted_waits[$index] ?? 0),
				);
			}
		} else {
			$message = sanitize_textarea_field(wp_unslash($_POST['message'] ?? ''));
			if ('' !== trim($message)) {
				$messages[] = array(
					'message' => MZW_WooCommerce::render_order_template($message, $order, str_replace('wc-', '', $status)),
					'wait'    => 0,
				);
			}
		}

		if (empty($messages)) {
			wp_send_json_error(array('message' => __('Informe ao menos uma mensagem para previsualizar.', 'miauzap-woocommerce')), 400);
		}
		$messages[0]['wait'] = 0;

		wp_send_json_success(array(
			'preview'  => $messages[0]['message'],
			'messages' => $messages,
		));
	}

	public function ajax_test_instance() {
		check_ajax_referer('mzw_admin', 'nonce');
		if (!current_user_can('manage_woocommerce')) {
			wp_send_json_error(array('message' => __('Permissao negada.', 'miauzap-woocommerce')), 403);
		}

		$key = sanitize_key($_POST['instance_key'] ?? '');
		$instance = $key ? MZW_Options::get_instance($key) : MZW_Options::pick_instance();
		$client = new MZW_API_Client();
		$response = $instance ? $client->status($instance) : $client->health();

		if (is_wp_error($response)) {
			wp_send_json_error(array('message' => $response->get_error_message()));
		}

		wp_send_json_success(array('response' => $response));
	}
}
