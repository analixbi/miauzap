(function ($) {
	'use strict';

	function setMessage($box, message, isError) {
		$box.find('.mzw-otp-message').text(message || '').toggleClass('is-error', !!isError);
	}

	function normalizePhoneInput(value) {
		var digits = String(value || '').replace(/\D+/g, '');
		if (digits.indexOf('55') === 0 && digits.length > 11) {
			digits = digits.substring(2);
		}
		return digits.substring(0, 11);
	}

	function formatPhoneInput(value) {
		var digits = normalizePhoneInput(value);
		if (digits.length <= 2) {
			return digits;
		}
		if (digits.length <= 10) {
			return '(' + digits.substring(0, 2) + ') ' + digits.substring(2, 6) + (digits.length > 6 ? '-' + digits.substring(6) : '');
		}
		return '(' + digits.substring(0, 2) + ') ' + digits.substring(2, 7) + '-' + digits.substring(7);
	}

	$(document).on('input', '.mzw-otp-phone', function () {
		this.value = formatPhoneInput(this.value);
	});

	$(document).on('click', '.mzw-send-otp', function () {
		var $box = $(this).closest('.mzw-otp-box');
		var $button = $(this);

		$button.prop('disabled', true);
		setMessage($box, 'Enviando...', false);

		$.post(MZWFrontend.ajaxUrl, {
			action: 'mzw_send_otp',
			nonce: MZWFrontend.nonce,
			phone: normalizePhoneInput($box.find('.mzw-otp-phone').val())
		}).done(function (response) {
			if (response.success) {
				setMessage($box, response.data.message, false);
				$box.find('.mzw-otp-code-wrap').prop('hidden', false);
				$box.find('.mzw-otp-code').trigger('focus');
			} else {
				setMessage($box, response.data && response.data.message ? response.data.message : 'Nao foi possivel enviar.', true);
			}
		}).fail(function () {
			setMessage($box, 'Falha ao enviar o codigo.', true);
		}).always(function () {
			$button.prop('disabled', false);
		});
	});

	$(document).on('click', '.mzw-verify-otp', function () {
		var $box = $(this).closest('.mzw-otp-box');
		var $button = $(this);

		$button.prop('disabled', true);
		setMessage($box, 'Validando...', false);

		$.post(MZWFrontend.ajaxUrl, {
			action: 'mzw_verify_otp',
			nonce: MZWFrontend.nonce,
			phone: normalizePhoneInput($box.find('.mzw-otp-phone').val()),
			code: $box.find('.mzw-otp-code').val()
		}).done(function (response) {
			if (response.success) {
				setMessage($box, response.data.message, false);
				window.location.href = response.data.redirect || window.location.href;
			} else {
				setMessage($box, response.data && response.data.message ? response.data.message : 'Codigo invalido.', true);
			}
		}).fail(function () {
			setMessage($box, 'Falha ao validar o codigo.', true);
		}).always(function () {
			$button.prop('disabled', false);
		});
	});
})(jQuery);
