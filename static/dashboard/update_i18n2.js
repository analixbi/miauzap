const fs = require('fs');

const path = 'c:/Users/Usuário/Desktop/Analix/miauzap/static/i18n.js';
let content = fs.readFileSync(path, 'utf8');

const newKeys = {
    pt: {
        "btn_cancel": "Cancelar",
        "btn_delete": "Deletar",
        "btn_save": "Salvar",
        "btn_generate_key": "<i class=\\"random icon\\"></i> Gerar Chave",
        "btn_show_key": "<i class=\\"eye icon\\"></i> Mostrar Chave",
        "btn_hide_key": "<i class=\\"eye slash icon\\"></i> Esconder Chave",
        
        "modal_chatwoot_enabled": "Ativar Integração Chatwoot",
        "modal_chatwoot_enabled_desc": "Ative ou desative a integração para esta instância.",
        "modal_chatwoot_conn_settings": "Configurações de Conexão",
        "modal_chatwoot_send_status": "Enviar Stories/Status",
        "modal_chatwoot_send_status_desc": "Se ativado, os status do WhatsApp serão enviados ao Chatwoot.",
        "modal_chatwoot_send_typing": "Sincronizar 'Digitando'",
        "modal_chatwoot_send_typing_desc": "Mostre o status 'digitando...' no WhatsApp ao escrever no Chatwoot.",
        "modal_chatwoot_url": "URL do Chatwoot <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_url_desc": "URL base da sua instância (ex: https://app.chatwoot.com)",
        "modal_chatwoot_account_id": "ID da Conta <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_account_id_desc": "ID da sua conta no Chatwoot",
        "modal_chatwoot_token": "Token da API <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_token_desc": "Token de acesso do Chatwoot",
        "modal_chatwoot_inbox_settings": "Configurações da Caixa de Entrada",
        "modal_chatwoot_inbox_name": "Nome da Caixa de Entrada",
        "modal_chatwoot_inbox_name_desc": "Nome da caixa no Chatwoot (padrão: nome da instância)",
        "modal_chatwoot_auto_create": "Criar Caixa Automaticamente",
        "modal_chatwoot_auto_create_desc": "Cria a caixa de entrada no Chatwoot caso não exista",
        "modal_chatwoot_msg_settings": "Configurações de Mensagem",
        "modal_chatwoot_sign_msg": "Assinar Mensagens",
        "modal_chatwoot_sign_msg_desc": "Adiciona a assinatura do atendente nas mensagens",
        "modal_chatwoot_sign_delimiter": "Separador de Assinatura",
        "modal_chatwoot_sign_delimiter_desc": "Separador usado (padrão: \\n\\n---\\n)",
        "modal_chatwoot_conv_settings": "Configurações de Conversa",
        "modal_chatwoot_reopen": "Reabrir Conversas",
        "modal_chatwoot_reopen_desc": "Reabre conversas automaticamente ao receber novas mensagens",
        "modal_chatwoot_pending": "Definir Conversas como Pendentes",
        "modal_chatwoot_pending_desc": "Novas conversas ficarão com status pendente",
        "modal_chatwoot_merge_brazil": "Mesclar Números Brasileiros",
        "modal_chatwoot_merge_brazil_desc": "Mescla contatos com variações do 9º dígito",
        "modal_chatwoot_import_groups": "Importar Mensagens de Grupo",
        "modal_chatwoot_import_groups_desc": "Importa mensagens de grupos do WhatsApp para o Chatwoot",
        "modal_chatwoot_webhook_info": "Informações do Webhook",
        "modal_chatwoot_webhook_url": "URL do Webhook (Leitura)",
        "modal_chatwoot_webhook_url_desc": "Configure esta URL na sua caixa de entrada no Chatwoot"
    },
    en: {
        "btn_cancel": "Cancel",
        "btn_delete": "Delete",
        "btn_save": "Save",
        "btn_generate_key": "<i class=\\"random icon\\"></i> Generate Random Key",
        "btn_show_key": "<i class=\\"eye icon\\"></i> Show Key",
        "btn_hide_key": "<i class=\\"eye slash icon\\"></i> Hide Key",

        "modal_chatwoot_enabled": "Enable Chatwoot Integration",
        "modal_chatwoot_enabled_desc": "Enable or disable Chatwoot integration for this instance.",
        "modal_chatwoot_conn_settings": "Connection Settings",
        "modal_chatwoot_send_status": "Send Status Stories",
        "modal_chatwoot_send_status_desc": "If enabled, status updates will be forwarded to Chatwoot.",
        "modal_chatwoot_send_typing": "Sync 'Typing' Presence",
        "modal_chatwoot_send_typing_desc": "Show 'typing...' status in WhatsApp when writing in Chatwoot.",
        "modal_chatwoot_url": "Chatwoot URL <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_url_desc": "Base URL of your instance (e.g., https://app.chatwoot.com)",
        "modal_chatwoot_account_id": "Account ID <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_account_id_desc": "Your Chatwoot account ID",
        "modal_chatwoot_token": "API Token <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_token_desc": "Chatwoot API access token",
        "modal_chatwoot_inbox_settings": "Inbox Settings",
        "modal_chatwoot_inbox_name": "Inbox Name",
        "modal_chatwoot_inbox_name_desc": "Name for the inbox in Chatwoot (defaults to instance name if empty)",
        "modal_chatwoot_auto_create": "Auto-create Inbox",
        "modal_chatwoot_auto_create_desc": "Automatically create inbox in Chatwoot if it doesn't exist",
        "modal_chatwoot_msg_settings": "Message Settings",
        "modal_chatwoot_sign_msg": "Sign Messages",
        "modal_chatwoot_sign_msg_desc": "Add agent signature to messages sent from Chatwoot",
        "modal_chatwoot_sign_delimiter": "Sign Delimiter",
        "modal_chatwoot_sign_delimiter_desc": "Delimiter for message signatures (default: \\n\\n---\\n)",
        "modal_chatwoot_conv_settings": "Conversation Settings",
        "modal_chatwoot_reopen": "Reopen Conversations",
        "modal_chatwoot_reopen_desc": "Automatically reopen closed conversations when new messages arrive",
        "modal_chatwoot_pending": "Set Conversations as Pending",
        "modal_chatwoot_pending_desc": "Set new conversations as pending status",
        "modal_chatwoot_merge_brazil": "Merge Brazilian Phone Numbers",
        "modal_chatwoot_merge_brazil_desc": "Merge contacts with Brazilian phone number variations (with/without 9th digit)",
        "modal_chatwoot_import_groups": "Import Group Messages",
        "modal_chatwoot_import_groups_desc": "Import messages from WhatsApp groups to Chatwoot",
        "modal_chatwoot_webhook_info": "Webhook Information",
        "modal_chatwoot_webhook_url": "Webhook URL (Read-only)",
        "modal_chatwoot_webhook_url_desc": "Configure this URL in your Chatwoot inbox settings to receive messages from agents"
    },
    es: {
        "btn_cancel": "Cancelar",
        "btn_delete": "Eliminar",
        "btn_save": "Guardar",
        "btn_generate_key": "<i class=\\"random icon\\"></i> Generar Clave Aleatoria",
        "btn_show_key": "<i class=\\"eye icon\\"></i> Mostrar Clave",
        "btn_hide_key": "<i class=\\"eye slash icon\\"></i> Ocultar Clave",

        "modal_chatwoot_enabled": "Habilitar Integración Chatwoot",
        "modal_chatwoot_enabled_desc": "Habilita o deshabilita la integración Chatwoot para esta instancia.",
        "modal_chatwoot_conn_settings": "Ajustes de Conexión",
        "modal_chatwoot_send_status": "Enviar Estados/Historias",
        "modal_chatwoot_send_status_desc": "Si está habilitado, los estados se enviarán a Chatwoot.",
        "modal_chatwoot_send_typing": "Sincronizar Presencia 'Escribiendo'",
        "modal_chatwoot_send_typing_desc": "Muestra el estado 'escribiendo...' en WhatsApp al escribir en Chatwoot.",
        "modal_chatwoot_url": "URL de Chatwoot <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_url_desc": "URL base de tu instancia (ej: https://app.chatwoot.com)",
        "modal_chatwoot_account_id": "ID de Cuenta <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_account_id_desc": "ID de tu cuenta de Chatwoot",
        "modal_chatwoot_token": "Token API <span style=\\"color: red;\\">*</span>",
        "modal_chatwoot_token_desc": "Token de acceso a la API de Chatwoot",
        "modal_chatwoot_inbox_settings": "Ajustes de la Bandeja",
        "modal_chatwoot_inbox_name": "Nombre de la Bandeja",
        "modal_chatwoot_inbox_name_desc": "Nombre de la bandeja en Chatwoot (por defecto: nombre de instancia)",
        "modal_chatwoot_auto_create": "Autocrear Bandeja",
        "modal_chatwoot_auto_create_desc": "Crea automáticamente la bandeja en Chatwoot si no existe",
        "modal_chatwoot_msg_settings": "Ajustes de Mensaje",
        "modal_chatwoot_sign_msg": "Firmar Mensajes",
        "modal_chatwoot_sign_msg_desc": "Añade la firma del agente a los mensajes",
        "modal_chatwoot_sign_delimiter": "Separador de Firma",
        "modal_chatwoot_sign_delimiter_desc": "Separador utilizado (por defecto: \\n\\n---\\n)",
        "modal_chatwoot_conv_settings": "Ajustes de Conversación",
        "modal_chatwoot_reopen": "Reabrir Conversaciones",
        "modal_chatwoot_reopen_desc": "Reabre conversaciones cerradas al recibir mensajes nuevos",
        "modal_chatwoot_pending": "Marcar como Pendientes",
        "modal_chatwoot_pending_desc": "Las nuevas conversaciones tendrán estado pendiente",
        "modal_chatwoot_merge_brazil": "Fusionar Números Brasileños",
        "modal_chatwoot_merge_brazil_desc": "Fusiona contactos con y sin el 9° dígito",
        "modal_chatwoot_import_groups": "Importar Mensajes de Grupos",
        "modal_chatwoot_import_groups_desc": "Importa mensajes de grupos de WhatsApp a Chatwoot",
        "modal_chatwoot_webhook_info": "Información del Webhook",
        "modal_chatwoot_webhook_url": "URL de Webhook (Solo Lectura)",
        "modal_chatwoot_webhook_url_desc": "Configura esta URL en Chatwoot para recibir mensajes de agentes"
    }
};

for (const [key, value] of Object.entries(newKeys.pt)) {
    content = content.replace(/(pt:\s*{[\s\S]*?)(},)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}
for (const [key, value] of Object.entries(newKeys.en)) {
    content = content.replace(/(en:\s*{[\s\S]*?)(},)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}
for (const [key, value] of Object.entries(newKeys.es)) {
    content = content.replace(/(es:\s*{[\s\S]*?)(}\s*};)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}

fs.writeFileSync(path, content, 'utf8');
console.log('Done');
