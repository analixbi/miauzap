(function ($) {
	'use strict';

	var focusedTemplate = null;

	function insertAtCursor($field, value) {
		var field = $field.get(0);
		if (!field) {
			return;
		}

		var start = field.selectionStart || 0;
		var end = field.selectionEnd || 0;
		var text = field.value;
		field.value = text.substring(0, start) + value + text.substring(end);
		field.focus();
		field.selectionStart = field.selectionEnd = start + value.length;
		$field.trigger('change');
	}

	function reindexFlow($flow) {
		var status = $flow.data('status');
		$flow.find('.mzw-flow-step').each(function (index) {
			var $step = $(this);
			var $wait = $step.find('.mzw-flow-wait');
			var $waitInput = $step.find('.mzw-flow-wait-input');
			$step.find('.mzw-flow-step-title').text('Mensagem ' + (index + 1));
			$step.find('.mzw-template').attr('name', 'order_templates[' + status + '][messages][' + index + '][message]');
			$waitInput.attr('name', 'order_templates[' + status + '][messages][' + index + '][wait]');
			$step.find('.mzw-remove-message').prop('hidden', index === 0);

			if (index === 0) {
				$wait.addClass('mzw-flow-wait-hidden');
				$waitInput.val('0');
			} else {
				$wait.removeClass('mzw-flow-wait-hidden');
				if (!$waitInput.val()) {
					$waitInput.val('0');
				}
			}
		});
	}

	function collectFlow($card) {
		var messages = [];
		var waits = [];
		$card.find('.mzw-flow-step').each(function (index) {
			var $step = $(this);
			messages.push($step.find('.mzw-template').val() || '');
			waits.push(index === 0 ? 0 : parseInt($step.find('.mzw-flow-wait-input').val(), 10) || 0);
		});

		return {
			messages: messages,
			waits: waits
		};
	}

	function createFlowStep(status, index) {
		var $step = $('<div/>', {
			'class': 'mzw-flow-step'
		});
		var $head = $('<div/>', {
			'class': 'mzw-flow-step-head'
		}).appendTo($step);

		$('<strong/>', {
			'class': 'mzw-flow-step-title'
		}).text('Mensagem ' + (index + 1)).appendTo($head);
		$('<button/>', {
			'class': 'button-link-delete mzw-remove-message',
			type: 'button'
		}).text('Remover').appendTo($head);

		var $wait = $('<label/>', {
			'class': 'mzw-flow-wait'
		}).appendTo($step);
		$wait.append(document.createTextNode('Aguardar antes desta mensagem '));
		$('<input/>', {
			'class': 'small-text mzw-flow-wait-input',
			min: 0,
			name: 'order_templates[' + status + '][messages][' + index + '][wait]',
			step: 1,
			type: 'number',
			value: 0
		}).appendTo($wait);
		$wait.append(document.createTextNode(' segundos'));

		$('<textarea/>', {
			'class': 'large-text code mzw-template',
			name: 'order_templates[' + status + '][messages][' + index + '][message]',
			rows: 7
		}).appendTo($step);

		return $step;
	}

	function setPreviewMessage(message, isError) {
		var $out = $('#mzw-preview-output');
		$out.empty();
		$('<div/>', {
			'class': 'mzw-whatsapp-bubble' + (isError ? ' is-error' : '')
		}).text(message).appendTo($out);
	}

	function renderPreviewMessages(messages) {
		var $out = $('#mzw-preview-output');
		var rendered = 0;
		$out.empty();

		(messages || []).forEach(function (item, index) {
			if (!item || !item.message) {
				return;
			}
			var wait = parseInt(item.wait, 10) || 0;

			if (index > 0 && wait > 0) {
				$('<div/>', {
					'class': 'mzw-preview-wait'
				}).text('+' + wait + 's').appendTo($out);
			}

			$('<div/>', {
				'class': 'mzw-whatsapp-bubble'
			}).text(item.message).appendTo($out);
			rendered += 1;
		});

		if (!rendered) {
			setPreviewMessage('Informe ao menos uma mensagem para previsualizar.', true);
		}
	}

	$(document).on('focus', '.mzw-template', function () {
		focusedTemplate = $(this);
	});

	$(document).on('click', '.mzw-var', function () {
		if (!focusedTemplate || !focusedTemplate.length) {
			focusedTemplate = $('.mzw-template:visible').first();
		}
		insertAtCursor(focusedTemplate, $(this).data('var'));
	});

	$(document).on('click', '#mzw-insert-note-var', function () {
		var keyword = ($('#mzw-note-keyword').val() || '').trim();
		var scope = $('#mzw-note-scope').val() || 'nota';
		if (!keyword) {
			window.alert('Informe a palavra-chave da anotacao, por exemplo Remessa_JET.');
			return;
		}
		if (!focusedTemplate || !focusedTemplate.length) {
			focusedTemplate = $('.mzw-template:visible').first();
		}
		insertAtCursor(focusedTemplate, '{' + scope + ':' + keyword + '}');
	});

	$(document).on('dragstart', '.mzw-var', function (event) {
		event.originalEvent.dataTransfer.setData('text/plain', $(this).data('var'));
	});

	$(document).on('dragover', '.mzw-template', function (event) {
		event.preventDefault();
	});

	$(document).on('drop', '.mzw-template', function (event) {
		event.preventDefault();
		insertAtCursor($(this), event.originalEvent.dataTransfer.getData('text/plain'));
	});

	$(document).on('click', '.mzw-add-message', function (event) {
		event.preventDefault();
		var $card = $(this).closest('.mzw-card');
		var $flow = $card.find('.mzw-flow');
		var status = $flow.data('status') || '';
		var $newStep = createFlowStep(status, $flow.find('.mzw-flow-step').length);

		$flow.append($newStep);
		reindexFlow($flow);
		$newStep.find('.mzw-template').trigger('focus');
	});

	$(document).on('click', '.mzw-remove-message', function () {
		var $flow = $(this).closest('.mzw-flow');
		if ($flow.find('.mzw-flow-step').length <= 1) {
			return;
		}

		$(this).closest('.mzw-flow-step').remove();
		reindexFlow($flow);
	});

	$(document).on('click', '.mzw-preview-template', function () {
		var $button = $(this);
		var $card = $button.closest('.mzw-card');
		var flow = collectFlow($card);
		var status = $button.data('status');
		var statusLabel = $button.data('status-label') || status;
		var orderId = $('#mzw-preview-order-id').val();
		var $status = $('#mzw-preview-status');

		$('.mzw-card').removeClass('mzw-card-active');
		$card.addClass('mzw-card-active');
		$status.text(statusLabel);
		setPreviewMessage('Carregando previsualizacao...');
		$.post(MZWAdmin.ajaxUrl, {
			action: 'mzw_preview_order_template',
			nonce: MZWAdmin.nonce,
			messages: flow.messages,
			waits: flow.waits,
			status: status,
			order_id: orderId
		}).done(function (response) {
			if (response.success) {
				renderPreviewMessages(response.data.messages || [{ message: response.data.preview, wait: 0 }]);
			} else {
				setPreviewMessage(response.data && response.data.message ? response.data.message : 'Nao foi possivel gerar a previsualizacao.', true);
			}
		}).fail(function () {
			setPreviewMessage('Falha ao gerar a previsualizacao.', true);
		});
	});

	$('#mzw-add-instance').on('click', function () {
		var $tbody = $('.mzw-instances tbody');
		var index = $tbody.find('tr').length;
		var row = [
			'<tr>',
			'<td><input type="checkbox" name="instances[' + index + '][enabled]" value="1" checked></td>',
			'<td><input type="hidden" name="instances[' + index + '][key]" value=""><input type="text" name="instances[' + index + '][label]" value="" placeholder="secundaria"></td>',
			'<td><input type="password" name="instances[' + index + '][token]" value="" placeholder="token" autocomplete="new-password"></td>',
			'<td><input type="number" min="1" name="instances[' + index + '][weight]" value="1" class="small-text"></td>',
			'<td><button type="button" class="button mzw-test-instance">Testar</button> <button type="button" class="button mzw-remove-row">Remover</button></td>',
			'</tr>'
		].join('');
		$tbody.append(row);
	});

	$(document).on('click', '.mzw-remove-row', function () {
		if ($('.mzw-instances tbody tr').length > 1) {
			$(this).closest('tr').remove();
		} else {
			$(this).closest('tr').find('input[type="text"], input[type="password"]').val('');
		}
	});

	$(document).on('click', '.mzw-test-instance', function () {
		var $row = $(this).closest('tr');
		var key = $row.find('input[name$="[key]"]').val();
		if (!key) {
			window.alert('Salve a instancia antes de testar.');
			return;
		}

		$.post(MZWAdmin.ajaxUrl, {
			action: 'mzw_test_instance',
			nonce: MZWAdmin.nonce,
			instance_key: key
		}).done(function (response) {
			if (response.success) {
				window.alert('Instancia respondeu com sucesso.');
			} else {
				window.alert(response.data && response.data.message ? response.data.message : 'Falha ao testar instancia.');
			}
		}).fail(function () {
			window.alert('Falha ao testar instancia.');
		});
	});
})(jQuery);
