<?php
/**
 * Plugin Name: Miauzap para WooCommerce
 * Description: Integracao WooCommerce com a API Miauzap para login por WhatsApp, mensagens de pedidos, campanhas, fila e rotacao de instancias.
 * Version: 0.1.8
 * Author: Analix BI
 * Requires PHP: 7.4
 * Requires at least: 6.0
 * WC requires at least: 7.0
 * Text Domain: miauzap-woocommerce
 */

if (!defined('ABSPATH')) {
	exit;
}

define('MZW_VERSION', '0.1.8');
define('MZW_FILE', __FILE__);
define('MZW_DIR', plugin_dir_path(__FILE__));
define('MZW_URL', plugin_dir_url(__FILE__));

require_once MZW_DIR . 'includes/class-mzw-options.php';
require_once MZW_DIR . 'includes/class-mzw-api-client.php';
require_once MZW_DIR . 'includes/class-mzw-queue.php';
require_once MZW_DIR . 'includes/class-mzw-woocommerce.php';
require_once MZW_DIR . 'includes/class-mzw-otp.php';
require_once MZW_DIR . 'includes/class-mzw-admin.php';
require_once MZW_DIR . 'includes/class-mzw-plugin.php';

register_activation_hook(__FILE__, array('MZW_Plugin', 'activate'));
register_deactivation_hook(__FILE__, array('MZW_Plugin', 'deactivate'));

add_action('plugins_loaded', array('MZW_Plugin', 'instance'));
