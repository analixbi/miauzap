const fs = require('fs');

const path = 'c:/Users/Usuário/Desktop/Analix/miauzap/static/i18n.js';
let content = fs.readFileSync(path, 'utf8');

const newKeys = {
    pt: {
        "config_s3_title": "Armazenamento S3",
        "config_s3_desc": "Configure armazenamento compatível com S3 para mídias.",
        "config_lang_title": "Idioma",
        "config_lang_desc": "Selecione o idioma das mensagens internas.",
        "config_proxy_title": "Configurações de Proxy",
        "config_proxy_desc": "Configure o proxy para as conexões.",
        "config_webhook_title": "Eventos de Webhook",
        "config_webhook_desc": "Configure a URL do webhook e eventos.",
        "config_history_title": "Histórico de Mensagens",
        "config_history_desc": "Configure a retenção do histórico.",
        "config_hmac_title": "Chave HMAC",
        "config_hmac_desc": "Configure a chave HMAC para segurança.",
        "config_chatwoot_title": "Integração Chatwoot",
        "config_chatwoot_desc": "Configure a integração com o Chatwoot.",

        "user_info_title": "Informações do Usuário",
        "user_info_desc": "Busque usuários por número.",
        "user_avatar_title": "Avatar do Usuário",
        "user_avatar_desc": "Obtenha a foto pelo número.",
        "user_contacts_title": "Meus Contatos",
        "user_contacts_desc": "Obtenha a lista de contatos.",

        "chat_send_text": "Enviar Texto",
        "chat_send_text_desc": "Envia uma mensagem de texto simples.",
        "chat_delete_msg": "Deletar Mensagem",
        "chat_delete_msg_desc": "Deleta uma mensagem enviada.",

        "groups_list_title": "Listar Grupos",
        "groups_list_desc": "Veja seus grupos do WhatsApp.",
        "groups_create_title": "Criar Grupo",
        "groups_create_desc": "Crie um novo grupo no WhatsApp.",
        "groups_join_title": "Entrar em Grupo",
        "groups_join_desc": "Entre via link de convite.",

        "dev_api_title": "API Playground",
        "dev_api_desc": "Teste os endpoints da API.",

        "badge_config": "Config",
        "badge_user": "Usuário",
        "badge_chat": "Chat",
        "badge_groups": "Grupos",
        "badge_dev": "Dev",

        "div_config": "Configuração",
        "div_user": "Usuário",
        "div_chat": "Chat",
        "div_groups": "Grupos",
        "div_dev": "Desenvolvedor",

        "lbl_admin_mode": "<i class=\\"shield icon\\"></i> Modo Administrador",
        "lbl_admin_mode_desc": "Insira o token de Admin Global para gerenciar todas as instâncias.",
        "btn_back_to_instances": "<i class=\\"left arrow icon\\"></i> Voltar às Instâncias",

        "groups_title": "<i class=\\"users icon\\"></i><div class=\\"content\\">Gerenciador de Grupos<div class=\\"sub header\\">Gerencie seus grupos</div></div>",
        "btn_back_dashboard": "<i class=\\"left arrow icon\\"></i> Voltar ao Dashboard",
        "search_groups": "Buscar grupos por nome...",
        "filter_all_groups": "Todos os Grupos",
        "filter_normal": "Grupos Normais",
        "filter_community": "Comunidades",
        "filter_community_group": "Grupos de Comunidade",
        "btn_refresh": "<i class=\\"refresh icon\\"></i> Atualizar",
        "btn_create_group": "<i class=\\"plus icon\\"></i> Criar Grupo",
        "loading_groups": "Carregando grupos...",
        "no_groups": "<i class=\\"users icon\\"></i>Nenhum grupo encontrado",
        "btn_create_first_group": "Crie seu primeiro grupo",

        "modal_pair_title": "Obtendo Código",
        "modal_pair_info": "Como conectar?",
        "modal_pair_li1": "Abra seu Whatsapp",
        "modal_pair_li2": "Aparelhos conectados",
        "modal_pair_li3": "Conectar com código",
        "modal_pair_phone": "Telefone",
        "modal_pair_phone_plc": "Digite seu número de telefone",
        "modal_pair_phone_desc": "Pressione Enter para enviar",

        "modal_webhook_title": "Webhook",
        "modal_webhook_events": "Eventos de Webhook",
        "modal_webhook_url": "URL do Webhook",
        "modal_webhook_url_plc": "Digite a URL do webhook",
        "btn_close": "Fechar",
        "btn_set": "Salvar<i class=\\"checkmark icon\\"></i>",

        "modal_s3_title": "Configuração S3",
        "modal_s3_info": "<div class=\\"header\\">Armazenamento S3</div><p>Configure armazenamento para mídias.</p>",
        "modal_s3_endpoint": "S3 Endpoint URL",
        "modal_s3_endpoint_plc": "https://s3.amazonaws.com ou outro",
        "modal_s3_endpoint_desc": "Para AWS use o padrão. Para MinIO use a URL customizada.",

        "modal_chatwoot_title": "Integração Chatwoot",
        "modal_chatwoot_info": "<div class=\\"header\\">Chatwoot CRM</div><p>Configure a integração de duas vias.</p>",
        
        "history_title": "<i class=\\"history icon\\"></i> Histórico",
        "history_select_inst": "Selecione uma instância...",
        "history_no_chat": "Selecione uma instância para ver as conversas",
        "history_viewer": "Visualizador de Histórico",
        "history_select_chat": "Selecione uma conversa na esquerda",
        "history_login_req": "<i class=\\"lock icon\\"></i> Autenticação Obrigatória",
        "history_login_desc": "Esta página requer um token de acesso para continuar.",
        "history_login_btn": "Entrar",
        "history_cancel_btn": "Cancelar",
        "history_token_plc": "Digite seu token"
    },
    en: {
        "config_s3_title": "S3 Storage",
        "config_s3_desc": "Configure S3-compatible storage for media files.",
        "config_lang_title": "Language",
        "config_lang_desc": "Select the language used for internal API messages.",
        "config_proxy_title": "Proxy Settings",
        "config_proxy_desc": "Configure proxy for WhatsApp connections.",
        "config_webhook_title": "Webhook Events",
        "config_webhook_desc": "Configure webhook URL and event subscriptions.",
        "config_history_title": "Message History",
        "config_history_desc": "Configure message history retention logic.",
        "config_hmac_title": "HMAC Key",
        "config_hmac_desc": "Configure HMAC key for webhook security.",
        "config_chatwoot_title": "Chatwoot Integration",
        "config_chatwoot_desc": "Configure Chatwoot CRM integration for WhatsApp.",

        "user_info_title": "User Info",
        "user_info_desc": "Search users by phone number.",
        "user_avatar_title": "User Avatar",
        "user_avatar_desc": "Get user avatar by phone number.",
        "user_contacts_title": "Get Contacts",
        "user_contacts_desc": "Get all contacts.",

        "chat_send_text": "Send Text Message",
        "chat_send_text_desc": "Sends a text message.",
        "chat_delete_msg": "Delete Message",
        "chat_delete_msg_desc": "Deletes your sent message.",

        "groups_list_title": "List Groups",
        "groups_list_desc": "View all your WhatsApp groups.",
        "groups_create_title": "Create Group",
        "groups_create_desc": "Create a new WhatsApp group.",
        "groups_join_title": "Join Group",
        "groups_join_desc": "Join a group using an invite link.",

        "dev_api_title": "API Playground",
        "dev_api_desc": "Test API endpoints and preview messages.",

        "badge_config": "Config",
        "badge_user": "User",
        "badge_chat": "Chat",
        "badge_groups": "Groups",
        "badge_dev": "Dev",

        "div_config": "Configuration",
        "div_user": "User",
        "div_chat": "Chat",
        "div_groups": "Groups",
        "div_dev": "Developer",

        "lbl_admin_mode": "<i class=\\"shield icon\\"></i> Admin Mode",
        "lbl_admin_mode_desc": "Please enter your Global Admin token in the field to manage all instances.",
        "btn_back_to_instances": "<i class=\\"left arrow icon\\"></i> Back to Instances",

        "groups_title": "<i class=\\"users icon\\"></i><div class=\\"content\\">Groups Manager<div class=\\"sub header\\">Manage your WhatsApp groups</div></div>",
        "btn_back_dashboard": "<i class=\\"left arrow icon\\"></i> Back to Dashboard",
        "search_groups": "Search groups by name or description...",
        "filter_all_groups": "All Groups",
        "filter_normal": "Normal Groups",
        "filter_community": "Communities",
        "filter_community_group": "Community Groups",
        "btn_refresh": "<i class=\\"refresh icon\\"></i> Refresh",
        "btn_create_group": "<i class=\\"plus icon\\"></i> Create Group",
        "loading_groups": "Loading groups...",
        "no_groups": "<i class=\\"users icon\\"></i>No groups found",
        "btn_create_first_group": "Create your first group",

        "modal_pair_title": "Getting Pair Code",
        "modal_pair_info": "How to pair?",
        "modal_pair_li1": "Open your Whatsapp",
        "modal_pair_li2": "Link a device",
        "modal_pair_li3": "Link with pair code",
        "modal_pair_phone": "Phone",
        "modal_pair_phone_plc": "Type your phone number",
        "modal_pair_phone_desc": "Press Enter to submit",

        "modal_webhook_title": "Webhook",
        "modal_webhook_events": "Webhook Events",
        "modal_webhook_url": "Webhook URL",
        "modal_webhook_url_plc": "Type webhook URL",
        "btn_close": "Close",
        "btn_set": "Set<i class=\\"checkmark icon\\"></i>",

        "modal_s3_title": "S3 Configuration",
        "modal_s3_info": "<div class=\\"header\\">S3 Storage Configuration</div><p>Configure S3-compatible storage for media files.</p>",
        "modal_s3_endpoint": "S3 Endpoint URL",
        "modal_s3_endpoint_plc": "https://s3.amazonaws.com or custom endpoint",
        "modal_s3_endpoint_desc": "For AWS S3, use the default. For MinIO/others, use your custom endpoint.",

        "modal_chatwoot_title": "Chatwoot Integration",
        "modal_chatwoot_info": "<div class=\\"header\\">Chatwoot CRM</div><p>Configure two-way integration.</p>",
        
        "history_title": "<i class=\\"history icon\\"></i> Chat History",
        "history_select_inst": "Select an instance...",
        "history_no_chat": "Select an instance to view chats",
        "history_viewer": "Chat History Viewer",
        "history_select_chat": "Select a chat on the left to see the entries",
        "history_login_req": "<i class=\\"lock icon\\"></i> Authentication Required",
        "history_login_desc": "This page requires authentication. Please enter your token to continue.",
        "history_login_btn": "Login",
        "history_cancel_btn": "Cancel",
        "history_token_plc": "Enter your user token or admin token"
    },
    es: {
        "config_s3_title": "Almacenamiento S3",
        "config_s3_desc": "Configura almacenamiento compatible con S3 para medios.",
        "config_lang_title": "Idioma",
        "config_lang_desc": "Selecciona el idioma para mensajes internos de la API.",
        "config_proxy_title": "Ajustes de Proxy",
        "config_proxy_desc": "Configura un proxy para las conexiones de WhatsApp.",
        "config_webhook_title": "Eventos Webhook",
        "config_webhook_desc": "Configura la URL y los eventos del webhook.",
        "config_history_title": "Historial de Mensajes",
        "config_history_desc": "Configura la lógica de retención del historial.",
        "config_hmac_title": "Clave HMAC",
        "config_hmac_desc": "Configura la clave HMAC para seguridad.",
        "config_chatwoot_title": "Integración Chatwoot",
        "config_chatwoot_desc": "Configura la integración CRM Chatwoot.",

        "user_info_title": "Info del Usuario",
        "user_info_desc": "Busca usuarios por número de teléfono.",
        "user_avatar_title": "Avatar de Usuario",
        "user_avatar_desc": "Obtén el avatar por número.",
        "user_contacts_title": "Mis Contactos",
        "user_contacts_desc": "Obtén la lista de contactos.",

        "chat_send_text": "Enviar Texto",
        "chat_send_text_desc": "Envía un mensaje de texto.",
        "chat_delete_msg": "Borrar Mensaje",
        "chat_delete_msg_desc": "Borra un mensaje enviado.",

        "groups_list_title": "Listar Grupos",
        "groups_list_desc": "Ve todos tus grupos de WhatsApp.",
        "groups_create_title": "Crear Grupo",
        "groups_create_desc": "Crea un nuevo grupo de WhatsApp.",
        "groups_join_title": "Unirse a Grupo",
        "groups_join_desc": "Únete a un grupo con un enlace.",

        "dev_api_title": "API Playground",
        "dev_api_desc": "Prueba endpoints y previsualiza mensajes.",

        "badge_config": "Config",
        "badge_user": "Usuario",
        "badge_chat": "Chat",
        "badge_groups": "Grupos",
        "badge_dev": "Dev",

        "div_config": "Configuración",
        "div_user": "Usuario",
        "div_chat": "Chat",
        "div_groups": "Grupos",
        "div_dev": "Desarrollador",

        "lbl_admin_mode": "<i class=\\"shield icon\\"></i> Modo Administrador",
        "lbl_admin_mode_desc": "Ingresa el token de Admin Global para gestionar todo.",
        "btn_back_to_instances": "<i class=\\"left arrow icon\\"></i> Volver a Instancias",

        "groups_title": "<i class=\\"users icon\\"></i><div class=\\"content\\">Gestor de Grupos<div class=\\"sub header\\">Gestiona tus grupos</div></div>",
        "btn_back_dashboard": "<i class=\\"left arrow icon\\"></i> Volver al Dashboard",
        "search_groups": "Buscar grupos por nombre...",
        "filter_all_groups": "Todos los Grupos",
        "filter_normal": "Grupos Normales",
        "filter_community": "Comunidades",
        "filter_community_group": "Grupos de Comunidad",
        "btn_refresh": "<i class=\\"refresh icon\\"></i> Actualizar",
        "btn_create_group": "<i class=\\"plus icon\\"></i> Crear Grupo",
        "loading_groups": "Cargando grupos...",
        "no_groups": "<i class=\\"users icon\\"></i>No hay grupos",
        "btn_create_first_group": "Crea tu primer grupo",

        "modal_pair_title": "Obteniendo Código",
        "modal_pair_info": "¿Cómo conectar?",
        "modal_pair_li1": "Abre tu Whatsapp",
        "modal_pair_li2": "Dispositivos vinculados",
        "modal_pair_li3": "Vincular con código",
        "modal_pair_phone": "Teléfono",
        "modal_pair_phone_plc": "Escribe tu número",
        "modal_pair_phone_desc": "Presiona Enter para enviar",

        "modal_webhook_title": "Webhook",
        "modal_webhook_events": "Eventos de Webhook",
        "modal_webhook_url": "URL de Webhook",
        "modal_webhook_url_plc": "Escribe la URL del webhook",
        "btn_close": "Cerrar",
        "btn_set": "Guardar<i class=\\"checkmark icon\\"></i>",

        "modal_s3_title": "Configuración S3",
        "modal_s3_info": "<div class=\\"header\\">Almacenamiento S3</div><p>Configura almacenamiento para medios.</p>",
        "modal_s3_endpoint": "S3 Endpoint URL",
        "modal_s3_endpoint_plc": "https://s3.amazonaws.com u otro",
        "modal_s3_endpoint_desc": "Para AWS usa el valor por defecto.",

        "modal_chatwoot_title": "Integración Chatwoot",
        "modal_chatwoot_info": "<div class=\\"header\\">Chatwoot CRM</div><p>Configura la integración bidireccional.</p>",
        
        "history_title": "<i class=\\"history icon\\"></i> Historial",
        "history_select_inst": "Selecciona una instancia...",
        "history_no_chat": "Selecciona una instancia para ver chats",
        "history_viewer": "Visor de Historial",
        "history_select_chat": "Selecciona un chat en la izquierda",
        "history_login_req": "<i class=\\"lock icon\\"></i> Autenticación Obligatoria",
        "history_login_desc": "Esta página requiere un token de acceso.",
        "history_login_btn": "Iniciar Sesión",
        "history_cancel_btn": "Cancelar",
        "history_token_plc": "Escribe tu token"
    }
};

// Replace pt translations
for (const [key, value] of Object.entries(newKeys.pt)) {
    content = content.replace(/(pt:\s*{[\s\S]*?)(},)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}
for (const [key, value] of Object.entries(newKeys.en)) {
    content = content.replace(/(en:\s*{[\s\S]*?)(},)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}
for (const [key, value] of Object.entries(newKeys.es)) {
    content = content.replace(/(es:\s*{[\s\S]*?)(}\s*};)/, `$1,\n        "${key}": ${JSON.stringify(value)}$2`);
}

// Add placeholder support to changeLanguage
content = content.replace(
    'el.innerHTML = translations[lang][key];\\n        }\\n    });',
    \`el.innerHTML = translations[lang][key];
        }
    });

    const placeholders = document.querySelectorAll("[data-i18n-placeholder]");
    placeholders.forEach((el) => {
        const key = el.getAttribute("data-i18n-placeholder");
        if (translations[lang] && translations[lang][key]) {
            el.placeholder = translations[lang][key];
        }
    });\`
);

fs.writeFileSync(path, content, 'utf8');
console.log('Done');
