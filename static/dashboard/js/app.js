// Miauzap Dashboard JavaScript
// Manages instances, dashboard UI, and API playground

let instances = [];
let currentInstance = null;
let currentEndpoint = 'text';
let currentUserRole = '';
let currentUserJID = '';
let pendingDeleteInstanceId = null;
let reconnectPollTimer = null;

// API Base URL - empty for relative paths or set to absolute URL
const baseUrl = "";
const returnAdminTokenKey = 'returnAdminToken';
const tempAdminTokenKey = 'tempAdminToken';

// Models for API Playground
const apiModels = {
  text: {
    method: 'POST',
    path: '/chat/send/text',
    body: {
      number: '5511999999999',
      text: 'Hello from Miauzap!'
    }
  },
  image: {
    method: 'POST',
    path: '/chat/send/image',
    body: {
      number: '5511999999999',
      image: 'https://placehold.co/600x400.png',
      caption: 'Look at this image!'
    }
  },
  video: {
    method: 'POST',
    path: '/chat/send/video',
    body: {
      number: '5511999999999',
      video: 'https://www.w3schools.com/html/mov_bbb.mp4',
      caption: 'Check out this video!'
    }
  },
  audio: {
    method: 'POST',
    path: '/chat/send/audio',
    body: {
      number: '5511999999999',
      audio: 'https://www.w3schools.com/html/horse.mp3'
    }
  },
  document: {
    method: 'POST',
    path: '/chat/send/document',
    body: {
      number: '5511999999999',
      document: 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf',
      fileName: 'document.pdf',
      caption: 'Here is your document'
    }
  },
  sticker: {
    method: 'POST',
    path: '/chat/send/sticker',
    body: {
      number: '5511999999999',
      sticker: 'https://raw.githubusercontent.com/mauview/mauview/master/mauview.webp'
    }
  },
  location: {
    method: 'POST',
    path: '/chat/send/location',
    body: {
      number: '5511999999999',
      latitude: -23.55052,
      longitude: -46.633308,
      title: 'São Paulo',
      address: 'Praça da Sé'
    }
  },
  contact: {
    method: 'POST',
    path: '/chat/send/contact',
    body: {
      number: '5511999999999',
      contactName: 'Miau',
      contactPhone: '5511988888888'
    }
  },
  carousel: {
    method: 'POST',
    path: '/chat/send/carousel',
    body: {
      number: '5511999999999',
      text: 'Confira as opções:',
      footer: 'Miauzap',
      cards: [
        {
          title: 'Opção 1',
          description: 'Descrição da opção 1',
          image: 'https://placehold.co/600x400.png',
          buttons: [
            { id: 'btn1', text: 'Selecionar' },
            { id: 'btn2', text: 'Ver mais' }
          ]
        },
        {
          title: 'Opção 2',
          description: 'Descrição da opção 2',
          image: 'https://placehold.co/600x400.png',
          buttons: [
            { id: 'btn3', text: 'Selecionar' }
          ]
        }
      ]
    }
  },
  pix: {
    method: 'POST',
    path: '/chat/send/pix',
    body: {
      number: '5511999999999',
      text: 'Efetue o pagamento via PIX:',
      footer: 'Miauzap',
      pix_key: 'sua-chave-pix',
      pix_name: 'Seu Nome',
      pix_city: 'Sua Cidade',
      amount: 10.50
    }
  },
  poll: {
    method: 'POST',
    path: '/chat/send/poll',
    body: {
      number: '5511999999999',
      name: 'Qual seu animal favorito?',
      options: ['Gato', 'Cachorro', 'Passarinho'],
      selectableCount: 1
    }
  },
  edit: {
    method: 'POST',
    path: '/chat/send/edit',
    body: {
      number: '5511999999999',
      messageID: 'ABC123XYZ',
      newMessage: 'Mensagem editada!'
    }
  },
  delete: {
    method: 'POST',
    path: '/chat/delete',
    body: {
      number: '5511999999999',
      messageID: 'ABC123XYZ'
    }
  },
  react: {
    method: 'POST',
    path: '/chat/react',
    body: {
      number: '5511999999999',
      messageID: 'ABC123XYZ',
      reaction: '❤️'
    }
  },
  archive: {
    method: 'POST',
    path: '/chat/archive',
    body: {
      number: '5511999999999',
      archive: true
    }
  },
  presence: {
    method: 'POST',
    path: '/chat/presence',
    body: {
      number: '5511999999999',
      presence: 'composing'
    }
  },
  markread: {
    method: 'POST',
    path: '/chat/markread',
    body: {
      number: '5511999999999',
      messageID: 'ABC123XYZ'
    }
  },
  history: {
    method: 'GET',
    path: '/chat/history',
    params: {
        number: '5511999999999',
        count: 50
    }
  },
  status: {
    method: 'GET',
    path: '/session/status'
  },
  qr: {
    method: 'GET',
    path: '/session/qr'
  },
  logout: {
    method: 'POST',
    path: '/session/logout'
  },
  user_info: {
    method: 'POST',
    path: '/user/info',
    body: {
      phone: ['5511999999999']
    }
  },
  user_avatar: {
    method: 'POST',
    path: '/user/avatar',
    body: {
      phone: '5511999999999'
    }
  },
  get_contacts: {
    method: 'GET',
    path: '/user/contacts'
  },
  create_group: {
    method: 'POST',
    path: '/group/create',
    body: {
      name: 'Novo Grupo Miauzap',
      participants: ['5511999999999', '5511988888888']
    }
  },
  group_info: {
    method: 'GET',
    path: '/group/info',
    params: {
        jid: '1234567890@g.us'
    }
  },
  group_invitelink: {
    method: 'GET',
    path: '/group/invitelink',
    params: {
        jid: '1234567890@g.us'
    }
  },
  group_photo: {
    method: 'POST',
    path: '/group/photo',
    body: {
      jid: '1234567890@g.us',
      image: 'https://placehold.co/600x400.png'
    }
  },
  group_name: {
    method: 'POST',
    path: '/group/name',
    body: {
      jid: '1234567890@g.us',
      name: 'Novo Nome do Grupo'
    }
  },
  group_topic: {
    method: 'POST',
    path: '/group/topic',
    body: {
      jid: '1234567890@g.us',
      topic: 'Novo tópico do grupo'
    }
  },
  group_description: {
    method: 'POST',
    path: '/group/description',
    body: {
      jid: '1234567890@g.us',
      description: 'Nova descrição do grupo'
    }
  },
  group_participants: {
    method: 'POST',
    path: '/group/participants',
    body: {
      jid: '1234567890@g.us',
      participants: ['5511999999999'],
      action: 'add'
    }
  },
  group_announce: {
    method: 'POST',
    path: '/group/announce',
    body: {
      jid: '1234567890@g.us',
      announce: true
    }
  },
  group_locked: {
    method: 'POST',
    path: '/group/locked',
    body: {
      jid: '1234567890@g.us',
      locked: true
    }
  },
  group_ephemeral: {
    method: 'POST',
    path: '/group/ephemeral',
    body: {
      jid: '1234567890@g.us',
      expiration: 86400
    }
  },
  group_join: {
    method: 'POST',
    path: '/group/join',
    body: {
      link: 'https://chat.whatsapp.com/B123XYZ'
    }
  },
  group_inviteinfo: {
    method: 'GET',
    path: '/group/inviteinfo',
    params: {
        link: 'https://chat.whatsapp.com/B123XYZ'
    }
  },
  group_leave: {
    method: 'POST',
    path: '/group/leave',
    body: {
      jid: '1234567890@g.us'
    }
  }
};

// --- Language/Translation Management ---

function updateDashboardLanguage() {
  // Update fixed elements
  const cards = [
      { id: 'sendTextMessage', model: apiModels.text },
      { id: 'deleteMessage', model: apiModels.delete },
      { id: 'groupsList', model: apiModels.group_info },
      { id: 'createGroup', model: apiModels.create_group },
      { id: 'joinGroup', model: apiModels.group_join }
  ];

  cards.forEach(card => {
      const el = document.getElementById(card.id);
      if (el && card.model) {
          // You could optionally update descriptions here if needed
      }
  });
}

// window.changeLanguage removed to avoid overwriting i18n.js function

function getCurrentUILanguage() {
  if (window.i18n && typeof window.i18n.getSavedLanguage === 'function') {
    return window.i18n.getSavedLanguage();
  }
  return localStorage.getItem('miauzap_lang') || sessionStorage.getItem('miauzap_lang_temp') || 'pt';
}

function applyCurrentUILanguage() {
  const langSelect = document.getElementById('langSelect');
  const lang = langSelect?.value || getCurrentUILanguage();

  if (typeof window.changeLanguage === 'function') {
    window.changeLanguage(lang);
  } else {
    localStorage.setItem('miauzap_lang', lang);
    sessionStorage.setItem('miauzap_lang_temp', lang);
  }
}

// --- DOM Ready ---

document.addEventListener('DOMContentLoaded', function () {
  const loginForm = document.getElementById('loginForm');
  const addInstanceForm = document.getElementById('addInstanceForm');
  const menulogout = document.getElementById('menulogout');
  const addInstanceButton = document.getElementById('addInstanceButton');
  const loginTokenInput = document.getElementById('loginToken');
  const loginAsAdminBtn = document.getElementById('loginAsAdminBtn');

  // Initialize Language
  const savedLang = getCurrentUILanguage();
  const langSelect = document.getElementById('langSelect');
  if (langSelect) langSelect.value = savedLang;
  applyCurrentUILanguage();

  // Login Logic
  if (loginForm) {
    loginForm.addEventListener('submit', function (e) {
      e.preventDefault();
      const token = loginTokenInput.value.trim();
      if (!token) {
        showError('Please enter an access token');
        return;
      }
      
      const role = $('input[name="loginRole"]:checked').val() || 'user';
      
      if (role === 'admin') {
        localStorage.setItem('adminToken', token);
        localStorage.removeItem('token');
        clearReturnAdminToken();
      } else {
        localStorage.setItem('token', token);
        localStorage.removeItem('adminToken');
        clearReturnAdminToken();
      }
      applyCurrentUILanguage();
      checkTokenAndLoad();
    });
  }

  if (menulogout) {
    menulogout.addEventListener('click', function (e) {
      e.preventDefault();
      localStorage.removeItem('token');
      localStorage.removeItem('adminToken');
      clearReturnAdminToken();
      location.reload();
    });
  }

  // Instance Management
  if (addInstanceButton) {
    addInstanceButton.addEventListener('click', function () {
      // Generate initial token for new instance
      const randomToken = generateRandomString(32);
      document.getElementById('instanceTokenInput').value = randomToken;
      $('#addInstanceModal').modal('show');
    });
  }

  if (addInstanceForm) {
    addInstanceForm.addEventListener('submit', function (e) {
      e.preventDefault();
      addInstance();
    });
  }

  document.getElementById('confirmDeleteButton')?.addEventListener('click', function () {
    deletePendingInstance();
  });
  document.getElementById('instanceReconnectButton')?.addEventListener('click', function () {
    reconnectCurrentInstance(this);
  });

  // Token Auto-load
  checkTokenAndLoad();

  // API Playground Events
  setupPlaygroundEvents();
  
  // Setup HMAC and other toggles in modals
  setupModalToggles();
});

// --- Authentication & Initialization ---

async function checkTokenAndLoad() {
  const token = localStorage.getItem('token');
  const adminToken = localStorage.getItem('adminToken');

  if (!token && !adminToken) {
    document.querySelectorAll('.hidden').forEach(el => el.classList.add('hidden'));
    document.querySelector('.logingrid').classList.remove('hidden');
    return;
  }

  document.querySelector('.logingrid').classList.add('hidden');
  document.querySelectorAll('.maingrid, .menu-header').forEach(el => el.classList.remove('hidden'));

  if (adminToken) {
    currentUserRole = 'Administrator';
    document.querySelectorAll('.admingrid').forEach(el => el.classList.remove('hidden'));
    document.querySelectorAll('.instance-only').forEach(el => el.classList.add('hidden'));
    document.querySelectorAll('.maingrid:not(.admingrid)').forEach(el => el.classList.add('hidden'));
    
    // Explicitly hide instance-specific views
    document.getElementById('mainDashboard')?.classList.add('hidden');
    document.getElementById('apiPlaygroundSection')?.classList.add('hidden');
    
    // Hide 'Back to Admin' button
    document.getElementById('backToAdminMenuBtn')?.classList.add('hidden');
    document.getElementById('heroBackToAdminBtn')?.classList.add('hidden');

    loadInstances();
  } else {
    currentUserRole = 'Instance User';
    document.querySelectorAll('.admingrid:not(.menu-header)').forEach(el => el.classList.add('hidden'));
    document.querySelectorAll('.instance-grid').forEach(el => el.classList.remove('hidden'));
    
    // Explicitly show instance dashboard
    document.getElementById('mainDashboard')?.classList.remove('hidden');
    
    // Show 'Back to Admin' only if we came from admin panel
    if (getReturnAdminToken()) {
        document.getElementById('backToAdminMenuBtn')?.classList.remove('hidden');
        document.getElementById('heroBackToAdminBtn')?.classList.remove('hidden');
        
        // Show the cards container if we need to return to it
        document.getElementById('instances-cards')?.parentElement?.classList.remove('hidden');
    } else {
        document.getElementById('backToAdminMenuBtn')?.classList.add('hidden');
        document.getElementById('heroBackToAdminBtn')?.classList.add('hidden');
    }

    loadSingleInstanceInfo();
  }

  document.getElementById('user-role-text').innerText = currentUserRole;
}

// --- API Calls (Core) ---

async function apiCall(endpoint, method = 'GET', body = null, customToken = null) {
  const token = customToken || localStorage.getItem('token') || localStorage.getItem('adminToken');
  const headers = new Headers();
  headers.append('token', token);
  headers.append('Authorization', token); // Send both to be safe
  
  if (body) headers.append('Content-Type', 'application/json');

  const options = {
    method: method,
    headers: headers
  };

  if (body) options.body = JSON.stringify(body);

  try {
    const response = await fetch(baseUrl + endpoint, options);
    const responseText = await response.text();
    let responseData = null;
    if (responseText) {
      try {
        responseData = JSON.parse(responseText);
      } catch (parseError) {
        responseData = null;
      }
    }

    if (!response.ok) {
        const errorMessage =
          responseData?.error ||
          responseData?.message ||
          responseData?.details ||
          responseData?.data?.error ||
          responseData?.data?.message ||
          responseText ||
          response.statusText ||
          'Request failed';
        return { code: response.status, error: errorMessage, details: responseData };
    }
    return responseData || { code: response.status, data: null };
  } catch (error) {
    console.error('API call error:', error);
    return { code: 500, error: 'Connection failed' };
  }
}

function normalizeChatwootUrl(value) {
  const trimmed = (value || '').trim().replace(/\/+$/, '');
  try {
    const parsed = new URL(trimmed);
    for (const marker of ['/app', '/api']) {
      const index = parsed.pathname.indexOf(marker);
      if (index >= 0) {
        parsed.pathname = parsed.pathname.slice(0, index);
        break;
      }
    }
    parsed.search = '';
    parsed.hash = '';
    return parsed.toString().replace(/\/+$/, '');
  } catch (error) {
    return trimmed;
  }
}

// --- Instance Management ---

async function loadInstances() {
  const loading = document.getElementById('loading');
  if (loading) loading.style.display = 'block';

  // For admin, the path is /admin/users
  const path = currentUserRole === 'Administrator' ? '/admin/users' : '/admin/users';
  const res = await apiCall(path, 'GET');
  
  if (loading) loading.style.display = 'none';

  if (res.code === 200 || Array.isArray(res)) {
    instances = Array.isArray(res) ? res : (res.data || []);
    renderInstancesTable();
  } else if (res.code === 401 || res.code === 403) {
    showError('Invalid admin token');
    localStorage.removeItem('adminToken');
    clearReturnAdminToken();
    location.reload();
  } else {
    showError(res.error || 'Failed to load instances');
  }
}

function renderInstancesTable() {
  const body = document.getElementById('instances-body');
  if (!body) return;
  body.innerHTML = '';

  instances.forEach(inst => {
    const id = inst.id || inst.ID || 'N/A';
    const name = inst.name || inst.Name || 'Unnamed';
    const token = inst.token || inst.Token || '';
    const connected = !!(inst.connected || inst.Connected);
    const loggedIn = !!(inst.loggedIn || inst.LoggedIn);
    const row = document.createElement('tr');

    const idCell = row.insertCell();
    idCell.textContent = id;

    const nameCell = row.insertCell();
    nameCell.textContent = name;

    const connectedCell = row.insertCell();
    connectedCell.className = connected ? 'positive' : 'negative';
    connectedCell.textContent = connected ? 'Yes' : 'No';

    const loggedInCell = row.insertCell();
    loggedInCell.className = loggedIn ? 'positive' : 'negative';
    loggedInCell.textContent = loggedIn ? 'Yes' : 'No';

    const actionsCell = row.insertCell();
    const actions = document.createElement('div');
    actions.className = 'ui small buttons';

    const manageButton = document.createElement('button');
    manageButton.className = 'ui button';
    manageButton.type = 'button';
    manageButton.textContent = 'Manage';
    manageButton.addEventListener('click', () => manageInstance(token));
    actions.appendChild(manageButton);

    const deleteButton = document.createElement('button');
    deleteButton.className = 'ui red button';
    deleteButton.type = 'button';
    deleteButton.textContent = 'Delete';
    deleteButton.addEventListener('click', () => confirmDeleteInstance(id));
    actions.appendChild(deleteButton);

    actionsCell.appendChild(actions);
    body.append(row);
  });
}

function manageInstance(token) {
  const currentAdminToken = localStorage.getItem('adminToken');
  if (currentAdminToken) {
      setReturnAdminToken(currentAdminToken);
  }
  localStorage.setItem('token', token);
  localStorage.removeItem('adminToken');
  location.reload();
}

function confirmDeleteInstance(id) {
  pendingDeleteInstanceId = id;
  $('#deleteInstanceModal').modal('show');
}

async function deletePendingInstance() {
  if (!pendingDeleteInstanceId) {
    showError('No instance selected');
    return;
  }

  const button = document.getElementById('confirmDeleteButton');
  if (button) button.classList.add('loading', 'disabled');

  const id = pendingDeleteInstanceId;
  const res = await apiCall(`/admin/users/${encodeURIComponent(id)}/full`, 'DELETE');

  if (button) button.classList.remove('loading', 'disabled');
  if (res.code === 200) {
    pendingDeleteInstanceId = null;
    $('#deleteInstanceModal').modal('hide');
    showSuccess('Instance deleted successfully');
    loadInstances();
  } else {
    showError(res.error || res.details || 'Failed to delete instance');
  }
}

async function addInstance() {
  const form = document.getElementById('addInstanceForm');
  const formData = new FormData(form);
  
  // Build the nested structure backend expects
  const data = {
      name: formData.get('name'),
      token: formData.get('token'),
      webhook: formData.get('webhook_url'), // Map webhook_url to webhook
      events: $('#webhookEventsInstance').dropdown('get value').join(','),
      history: parseInt(formData.get('history')) || 0,
      proxyConfig: {
          enabled: $('#addInstanceProxyToggle').checkbox('is checked'),
          proxyURL: formData.get('proxy_url')
      },
      s3Config: {
          enabled: $('#addInstanceS3Toggle').checkbox('is checked'),
          endpoint: formData.get('s3_endpoint'),
          access_key: formData.get('s3_access_key'),
          secret_key: formData.get('s3_secret_key'),
          bucket: formData.get('s3_bucket'),
          region: formData.get('s3_region'),
          public_url: formData.get('s3_public_url'),
          media_delivery: $('#addInstanceS3MediaDelivery').dropdown('get value'),
          retention_days: parseInt(formData.get('s3_retention_days')) || 30,
          path_style: $('#addInstanceS3PathStyleToggle').checkbox('is checked')
      },
      hmacKey: $('#addInstanceHmacToggle').checkbox('is checked') ? formData.get('hmac_key') : ""
  };

  const res = await apiCall('/admin/users', 'POST', data);

  if (res.code === 201 || res.code === 200) {
    $('#addInstanceModal').modal('hide');
    form.reset();
    showSuccess('Instance created successfully');
    loadInstances();
  } else {
    showError(res.error || 'Failed to create instance');
  }
}

// --- Single Instance Dashboard ---

async function loadSingleInstanceInfo() {
  const res = await apiCall('/session/status', 'GET');
  if (res.code === 200) {
    currentInstance = res.data;
    renderInstanceDashboard();
  } else {
    // If token invalid, logout
    localStorage.removeItem('token');
    location.reload();
  }
}

function renderInstanceDashboard() {
  const dashboard = document.getElementById('mainDashboard');
  if (!dashboard) return;
  
  // Show all relevant widgets
  document.querySelectorAll('.widget').forEach(el => el.classList.remove('hidden'));

  // Pre-load configurations for modals
  loadChatwootConfig();
  loadS3Config();

  // Pre-fill forms based on currentInstance
  if (currentInstance) {
    if (document.getElementById('displayInstanceName')) {
        document.getElementById('displayInstanceName').innerText = currentInstance.name || 'Instance Dashboard';
    }
    if (document.getElementById('displayInstanceId')) {
        document.getElementById('displayInstanceId').innerText = currentInstance.id || currentInstance.ID || 'N/A';
    }
    if (document.getElementById('displayInstanceToken')) {
        document.getElementById('displayInstanceToken').innerText = localStorage.getItem('token') || 'N/A';
    }
    const statusDiv = document.getElementById('displayInstanceStatus');
    if (statusDiv) {
        const isConnected = !!(currentInstance.connected || currentInstance.Connected);
        const isLoggedIn = !!(currentInstance.loggedIn || currentInstance.LoggedIn);
        if (isConnected && isLoggedIn) {
          statusDiv.innerHTML = '<span class="ui green label" style="margin-left: 10px;">Conectado</span>';
        } else if (isConnected) {
          statusDiv.innerHTML = '<span class="ui orange label" style="margin-left: 10px;">Aguardando QR</span>';
        } else {
          statusDiv.innerHTML = '<span class="ui red label" style="margin-left: 10px;">Desconectado</span>';
        }
    }

    updateReconnectControls();

    if (currentInstance.webhook) {
        document.getElementById('webhookinput').value = currentInstance.webhook;
    }
    if (currentInstance.events) {
        $('#webhookEvents').dropdown('set selected', currentInstance.events.split(','));
    }
  }
}

function getCurrentInstanceEvents() {
  return String(currentInstance?.events || currentInstance?.Events || '')
    .split(',')
    .map(event => event.trim())
    .filter(Boolean);
}

function updateReconnectControls() {
  const button = document.getElementById('instanceReconnectButton');
  const panel = document.getElementById('instanceReconnectPanel');
  if (!button || !currentInstance) return;

  const isConnected = !!(currentInstance.connected || currentInstance.Connected);
  const isLoggedIn = !!(currentInstance.loggedIn || currentInstance.LoggedIn);

  if (isConnected && isLoggedIn) {
    button.classList.add('hidden');
    hideReconnectPanel();
    stopReconnectPolling();
  } else {
    button.classList.remove('hidden');
    if (isConnected && !isLoggedIn) {
      panel?.classList.remove('hidden');
      loadReconnectQr();
      startReconnectPolling();
    }
  }
}

async function reconnectCurrentInstance(button) {
  if (!localStorage.getItem('token')) {
    showError('Token da instancia nao encontrado');
    return;
  }

  setReconnectMessage('Iniciando reconexao e aguardando QR Code...');
  showReconnectPanel();

  if (button) {
    button.classList.add('loading', 'disabled');
    button.disabled = true;
  }

  const res = await apiCall('/session/connect', 'POST', {
    Subscribe: getCurrentInstanceEvents(),
    Immediate: true
  });

  if (button) {
    button.classList.remove('loading', 'disabled');
    button.disabled = false;
  }

  const alreadyConnected = res.code === 500 && String(res.error || '').toLowerCase().includes('already connected');
  const success = res.code === 200 || alreadyConnected;
  if (!success) {
    setReconnectMessage(res.error || res.details || 'Nao foi possivel iniciar a reconexao.');
    showError(res.error || res.details || 'Falha ao reconectar a instancia');
    return;
  }

  await loadSingleInstanceInfo();
  await loadReconnectQr();
  startReconnectPolling();
}

function showReconnectPanel() {
  document.getElementById('instanceReconnectPanel')?.classList.remove('hidden');
}

function hideReconnectPanel() {
  document.getElementById('instanceReconnectPanel')?.classList.add('hidden');
  const img = document.getElementById('instanceQrImage');
  const placeholder = document.getElementById('instanceQrPlaceholder');
  if (img) {
    img.removeAttribute('src');
    img.style.display = 'none';
  }
  if (placeholder) placeholder.style.display = 'block';
}

function setReconnectMessage(message) {
  const messageEl = document.getElementById('instanceQrMessage');
  if (messageEl) messageEl.textContent = message;
}

async function loadReconnectQr() {
  const qrImage = document.getElementById('instanceQrImage');
  const placeholder = document.getElementById('instanceQrPlaceholder');
  const res = await apiCall('/session/qr', 'GET');

  if (res.code === 200 && res.data) {
    const qrCode = res.data.QRCode || res.data.qrcode || res.data.qrCodeBase64 || '';
    if (qrCode && qrImage) {
      qrImage.src = qrCode;
      qrImage.style.display = 'block';
      if (placeholder) placeholder.style.display = 'none';
      setReconnectMessage('Escaneie o QR Code no WhatsApp para concluir a reconexao.');
      return true;
    }
  }

  if (res.error && String(res.error).toLowerCase().includes('already logged in')) {
    setReconnectMessage('Instancia reconectada com sucesso.');
    hideReconnectPanel();
    stopReconnectPolling();
    await loadSingleInstanceInfo();
    return true;
  }

  setReconnectMessage(res.error || 'Aguardando o QR Code ser gerado...');
  if (qrImage) qrImage.style.display = 'none';
  if (placeholder) placeholder.style.display = 'block';
  return false;
}

function startReconnectPolling() {
  if (reconnectPollTimer) return;
  reconnectPollTimer = window.setInterval(async () => {
    const status = await apiCall('/session/status', 'GET');
    if (status.code === 200 && status.data) {
      currentInstance = status.data;
      renderInstanceDashboard();
      const isConnected = !!(currentInstance.connected || currentInstance.Connected);
      const isLoggedIn = !!(currentInstance.loggedIn || currentInstance.LoggedIn);
      if (isConnected && isLoggedIn) {
        setReconnectMessage('Instancia reconectada com sucesso.');
        hideReconnectPanel();
        stopReconnectPolling();
        return;
      }
    }
    await loadReconnectQr();
  }, 3500);
}

function stopReconnectPolling() {
  if (reconnectPollTimer) {
    window.clearInterval(reconnectPollTimer);
    reconnectPollTimer = null;
  }
}

// Config Loaders (called before modal shows)
async function loadS3Config() {
  const res = await apiCall('/session/s3/config', 'GET');
  if (res.code === 200 && res.data) {
    const data = res.data;
    document.getElementById('s3Endpoint').value = data.endpoint || '';
    document.getElementById('s3AccessKey').value = data.access_key !== '***' ? data.access_key : '';
    document.getElementById('s3SecretKey').value = data.secret_key !== '***' ? data.secret_key : '';
    document.getElementById('s3Bucket').value = data.bucket || '';
    document.getElementById('s3Region').value = data.region || '';
    document.getElementById('s3PublicUrl').value = data.public_url || '';
    if(data.media_delivery) $('#s3MediaDelivery').dropdown('set selected', data.media_delivery);
  }
}

async function loadChatwootConfig() {
  const res = await apiCall('/chatwoot/config', 'GET');
  if (res.code === 200 && res.data) {
    const data = res.data;
    $('#chatwootEnabledToggle').checkbox(data.enabled ? 'check' : 'uncheck');
    if (document.getElementById('chatwootUrl')) document.getElementById('chatwootUrl').value = data.url || '';
    if (document.getElementById('chatwootToken')) document.getElementById('chatwootToken').value = data.token !== '***' ? data.token : '';
    if (document.getElementById('chatwootAccountId')) document.getElementById('chatwootAccountId').value = data.accountId || '';
    if (document.getElementById('chatwootInboxId')) document.getElementById('chatwootInboxId').value = data.inboxId || '';
    if (document.getElementById('chatwootInboxName')) document.getElementById('chatwootInboxName').value = data.inboxName || '';
    if (document.getElementById('chatwootExistingWebhookUrl')) document.getElementById('chatwootExistingWebhookUrl').value = data.webhookUrl || '';
    if (document.getElementById('chatwootWebhookSecret')) {
        document.getElementById('chatwootWebhookSecret').value = data.webhookSecret && data.webhookSecret !== '********' ? data.webhookSecret : '';
        document.getElementById('chatwootWebhookSecret').placeholder = data.webhookSecret === '********' ? 'Segredo salvo. Preencha apenas se quiser alterar.' : 'Cole o segredo do webhook';
    }
    
    // Toggles
    $('#chatwootSignMsgToggle').checkbox(data.signMsg ? 'check' : 'uncheck');
    if (document.getElementById('chatwootSignDelimiter')) document.getElementById('chatwootSignDelimiter').value = data.signDelimiter || '';
    $('#chatwootReopenToggle').checkbox(data.reopenConversation ? 'check' : 'uncheck');
    $('#chatwootPendingToggle').checkbox(data.conversationPending ? 'check' : 'uncheck');
    $('#chatwootMergeBrazilToggle').checkbox(data.mergeBrazilContacts ? 'check' : 'uncheck');
    $('#chatwootImportGroupsToggle').checkbox(data.importGroups ? 'check' : 'uncheck');
    $('#chatwootSendStatusStoriesToggle').checkbox(data.sendStatusStories ? 'check' : 'uncheck');
    $('#chatwootSendTypingToggle').checkbox(data.sendTyping !== false ? 'check' : 'uncheck');
    $('#chatwootSendReadReceiptsToggle').checkbox(data.sendReadReceipts ? 'check' : 'uncheck');
    
    if (document.getElementById('chatwootWebhookUrl')) {
        document.getElementById('chatwootWebhookUrl').value = data.webhookUrl || '';
    }
  }
}

async function saveChatwootConfig() {
    const inboxIdRaw = document.getElementById('chatwootInboxId')?.value?.trim() || '';
    const inboxId = inboxIdRaw ? parseInt(inboxIdRaw, 10) : 0;
    if (inboxIdRaw && (!Number.isInteger(inboxId) || inboxId <= 0)) {
        showError('Informe um ID de caixa de entrada valido.');
        return;
    }

    const chatwootUrl = normalizeChatwootUrl(document.getElementById('chatwootUrl').value);
    document.getElementById('chatwootUrl').value = chatwootUrl;
    const existingWebhookUrl = document.getElementById('chatwootExistingWebhookUrl')?.value?.trim() || '';

    const data = {
        enabled: $('#chatwootEnabledToggle').checkbox('is checked'),
        url: chatwootUrl,
        token: document.getElementById('chatwootToken').value,
        accountId: document.getElementById('chatwootAccountId').value,
        inboxId: inboxId,
        nameInbox: document.getElementById('chatwootInboxName').value,
        webhookUrl: existingWebhookUrl,
        webhookSecret: document.getElementById('chatwootWebhookSecret')?.value?.trim() || '',
        signMsg: $('#chatwootSignMsgToggle').checkbox('is checked'),
        signDelimiter: document.getElementById('chatwootSignDelimiter').value,
        reopenConversation: $('#chatwootReopenToggle').checkbox('is checked'),
        conversationPending: $('#chatwootPendingToggle').checkbox('is checked'),
        mergeBrazilContacts: $('#chatwootMergeBrazilToggle').checkbox('is checked'),
        importGroups: $('#chatwootImportGroupsToggle').checkbox('is checked'),
        sendStatusStories: $('#chatwootSendStatusStoriesToggle').checkbox('is checked'),
        sendTyping: $('#chatwootSendTypingToggle').checkbox('is checked'),
        sendReadReceipts: $('#chatwootSendReadReceiptsToggle').checkbox('is checked'),
        autoCreate: $('#chatwootAutoCreateToggle').checkbox('is checked')
    };

    const res = await apiCall('/chatwoot/config', 'POST', data);
    if (res.code === 200) {
        const warning = res.data?.setupWarning;
        if (warning) {
            alert(warning);
            if (document.getElementById('chatwootWebhookUrl')) {
                document.getElementById('chatwootWebhookUrl').value = res.data?.webhookUrl || '';
            }
            if (document.getElementById('chatwootExistingWebhookUrl')) {
                document.getElementById('chatwootExistingWebhookUrl').value = res.data?.webhookUrl || '';
            }
        } else {
            showSuccess('Chatwoot configuration saved');
            $('#modalChatwootConfig').modal('hide');
        }
    } else {
        showError(res.error || 'Failed to save Chatwoot configuration');
    }
}

async function deleteChatwootConfig() {
    if (!confirm('Are you sure you want to delete Chatwoot configuration?')) return;
    
    const data = { enabled: false };
    const res = await apiCall('/chatwoot/config', 'POST', data);
    if (res.code === 200) {
        showSuccess('Chatwoot configuration deleted');
        $('#modalChatwootConfig').modal('hide');
    } else {
        showError(res.error || 'Failed to delete configuration');
    }
}

async function saveS3Config() {
    const data = {
        endpoint: document.getElementById('s3Endpoint').value,
        access_key: document.getElementById('s3AccessKey').value,
        secret_key: document.getElementById('s3SecretKey').value,
        bucket: document.getElementById('s3Bucket').value,
        region: document.getElementById('s3Region').value,
        public_url: document.getElementById('s3PublicUrl').value,
        media_delivery: $('#s3MediaDelivery').dropdown('get value'),
        enabled: true
    };

    const res = await apiCall('/session/s3/config', 'POST', data);
    if (res.code === 200) {
        showSuccess('S3 configuration saved');
        $('#modalS3Config').modal('hide');
    } else {
        showError(res.error || 'Failed to save S3 configuration');
    }
}


// --- API Playground Setup ---

function setupPlaygroundEvents() {
  const endpointItems = document.querySelectorAll('#apiEndpointsMenu .item[data-endpoint]');
  const sendBtn = document.getElementById('apiSendTestBtn');
  const jsonTextArea = document.getElementById('apiTestJson');
  
  endpointItems.forEach(item => {
    item.addEventListener('click', function() {
      endpointItems.forEach(i => i.classList.remove('active'));
      this.classList.add('active');
      currentEndpoint = this.dataset.endpoint;
      updatePlaygroundUI();
    });
  });

  if (sendBtn) {
    sendBtn.addEventListener('click', async function() {
      try {
        const body = JSON.parse(jsonTextArea.value);
        const model = apiModels[currentEndpoint];
        
        // Show loading in button
        sendBtn.classList.add('loading');
        
        let endpoint = model.path;
        let method = model.method;
        
        // Handle GET with params if needed
        if (method === 'GET' && body) {
            const params = new URLSearchParams(body).toString();
            endpoint += '?' + params;
        }

        const res = await apiCall(endpoint, method, method === 'GET' ? null : body);
        
        sendBtn.classList.remove('loading');
        displayPlaygroundResponse(res);
      } catch (err) {
        showError('Invalid JSON format');
      }
    });
  }

  // Copy buttons
  document.getElementById('copyJsonBtn')?.addEventListener('click', () => {
    copyToClipboard(jsonTextArea.value);
  });

  document.getElementById('copyCurlBtn')?.addEventListener('click', () => {
    const curl = document.getElementById('apiCurlPreview').innerText;
    copyToClipboard(curl);
  });

  // Tab switching (Body/Curl)
  $('.tabular.menu .item').tab();

  // Dashboard Cards Listeners
  document.getElementById('sendTextMessage')?.addEventListener('click', function () {
    currentEndpoint = 'text';
    activatePlayground();
  });

  document.getElementById('deleteMessage')?.addEventListener('click', function () {
    currentEndpoint = 'delete';
    activatePlayground();
  });

  document.getElementById('openApiPlayground')?.addEventListener('click', function () {
    activatePlayground();
  });
  
  // New Tiles
  const tileMap = {
    's3Config': { modal: 'modalS3Config', fetch: loadS3Config },
    'languageConfig': 'modalLanguageConfig',
    'proxyConfig': 'modalProxyConfig',
    'webhookConfig': 'modalSetWebhook',
    'historyConfig': 'modalHistoryConfig',
    'hmacConfig': 'modalHmacConfig',
    'chatwootConfig': { modal: 'modalChatwootConfig', fetch: loadChatwootConfig }
  };

  Object.entries(tileMap).forEach(([id, config]) => {
    document.getElementById(id)?.addEventListener('click', () => {
        if (typeof config === 'object' && config.fetch) {
            config.fetch().then(() => {
                $(`#${config.modal}`).modal('show');
            });
        } else {
            const modalId = typeof config === 'string' ? config : config.modal;
            $(`#${modalId}`).modal('show');
        }
    });
  });

  // Playground listeners for user tiles
  document.getElementById('userInfo')?.addEventListener('click', function () {
    currentEndpoint = 'user_info';
    activatePlayground();
  });
  document.getElementById('userAvatar')?.addEventListener('click', function () {
    currentEndpoint = 'user_avatar';
    activatePlayground();
  });
  document.getElementById('userContacts')?.addEventListener('click', function () {
    currentEndpoint = 'get_contacts';
    activatePlayground();
  });

  // Modal Action Listeners
  document.getElementById('saveChatwootConfig')?.addEventListener('click', saveChatwootConfig);
  document.getElementById('deleteChatwootConfig')?.addEventListener('click', deleteChatwootConfig);
  document.getElementById('saveS3Config')?.addEventListener('click', saveS3Config);
  document.getElementById('saveLanguageConfig')?.addEventListener('click', saveLanguageConfig);
  
  document.getElementById('copyChatwootWebhookUrl')?.addEventListener('click', () => {
      const url = document.getElementById('chatwootWebhookUrl').value;
      copyToClipboard(url);
  });

  // Groups.js handles groupsList, createGroup, joinGroup clicks
}

async function saveLanguageConfig() {
  const selectedLang = document.getElementById('instanceLanguageSelect')?.value || getCurrentUILanguage();

  if (typeof window.changeLanguage === 'function') {
    window.changeLanguage(selectedLang);
  }

  const instanceID = currentInstance?.id || currentInstance?.ID;
  if (instanceID) {
    const res = await apiCall(`/user/language/${encodeURIComponent(instanceID)}`, 'PUT', { language: selectedLang });
    if (res.code !== 200) {
      showError(res.error || 'Failed to save language');
      return;
    }
  }

  $('#modalLanguageConfig').modal('hide');
  showSuccess('Language saved');
}

function updatePlaygroundUI() {
  const model = apiModels[currentEndpoint];
  if (!model) return;

  document.getElementById('apiEndpointTitle').innerText = document.querySelector(`[data-endpoint="${currentEndpoint}"]`).innerText;
  
  const body = model.body || model.params || {};
  document.getElementById('apiTestJson').value = JSON.stringify(body, null, 2);
  
  // Model Schema
  document.getElementById('apiModelSchema').innerText = JSON.stringify(model, null, 2);
  
  updateCurlPreview();
  updateMessagePreview();
}

function updateCurlPreview() {
  const model = apiModels[currentEndpoint];
  const token = localStorage.getItem('token') || 'YOUR_TOKEN';
  const body = document.getElementById('apiTestJson').value;
  
  let curl = `curl -X ${model.method} "${window.location.origin}${model.path}" \\\n`;
  curl += `  -H "token: ${token}" \\\n`;
  if (model.method !== 'GET') {
    curl += `  -H "Content-Type: application/json" \\\n`;
    curl += `  -d '${body.replace(/'/g, "'\\''")}'`;
  }
  
  document.getElementById('apiCurlPreview').innerText = curl;
}

function updateMessagePreview() {
    const previewBubble = document.getElementById('previewBubble');
    const previewContent = document.getElementById('previewContent');
    const bodyText = document.getElementById('apiTestJson').value;
    
    try {
        const body = JSON.parse(bodyText);
        
        // Simple preview logic based on endpoint
        switch(currentEndpoint) {
            case 'text':
                previewContent.innerText = body.message || 'Hello!';
                break;
            case 'image':
                previewContent.innerHTML = `<img src="${body.image}" style="max-width: 100%; border-radius: 5px;"><br>${body.caption || ''}`;
                break;
            case 'carousel':
                let html = `<div>${body.text}</div><div style="display: flex; overflow-x: auto; gap: 10px; padding-top: 5px;">`;
                (body.cards || []).forEach(card => {
                    html += `<div style="min-width: 150px; background: white; border-radius: 5px; padding: 5px;">
                        <img src="${card.image}" style="width: 100%; border-radius: 3px;">
                        <div style="font-weight: bold; font-size: 0.9em;">${card.title}</div>
                        <div style="font-size: 0.8em; color: #666;">${card.description}</div>
                    </div>`;
                });
                html += '</div>';
                previewContent.innerHTML = html;
                break;
            default:
                previewContent.innerText = `Preview for ${currentEndpoint} message type`;
        }
    } catch(e) {}
}

function displayPlaygroundResponse(res) {
  const container = document.getElementById('apiResponse');
  const content = document.getElementById('apiResponseContent');
  
  container.classList.remove('hidden');
  content.innerText = JSON.stringify(res, null, 2);
  
  if (res.code >= 200 && res.code < 300) {
    container.className = 'ui message positive';
  } else {
    container.className = 'ui message negative';
  }
}

function activatePlayground() {
  document.getElementById('mainDashboard').classList.add('hidden');
  document.getElementById('apiPlaygroundSection').classList.remove('hidden');
  updatePlaygroundUI();
}

function goBackToList() {
    // If instance user, reload to dashboard. If admin, show list.
    if (currentUserRole === 'Administrator') {
      document.getElementById('apiPlaygroundSection')?.classList.add('hidden');
      document.getElementById('mainDashboard')?.classList.add('hidden');
      document.querySelectorAll('.admingrid').forEach(el => el.classList.remove('hidden'));
    } else {
       const tempAdmin = getReturnAdminToken();
       if (tempAdmin) {
           localStorage.setItem('adminToken', tempAdmin);
           localStorage.removeItem('token');
           clearReturnAdminToken();
       }
       location.reload();
    }
}

// --- Utilities ---

function setReturnAdminToken(token) {
  sessionStorage.setItem(tempAdminTokenKey, token);
  localStorage.setItem(returnAdminTokenKey, token);
}

function getReturnAdminToken() {
  return sessionStorage.getItem(tempAdminTokenKey) || localStorage.getItem(returnAdminTokenKey);
}

function clearReturnAdminToken() {
  sessionStorage.removeItem(tempAdminTokenKey);
  localStorage.removeItem(returnAdminTokenKey);
}

function showError(msg) {
  alert(msg); // Replace with a nice toast/modal later
}

function showSuccess(msg) {
  console.log('SUCCESS:', msg);
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text).then(() => {
    showSuccess('Copied to clipboard');
  });
}

function generateRandomString(length) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

function getLocalStorageItem(key) {
    return localStorage.getItem(key);
}

// --- Specific Modal Logics ---

function setupModalToggles() {
    // Webhook modal
    $('#webhookEvents').dropdown();
    
    // Add Instance toggles
    $('#addInstanceProxyToggle').checkbox({
        onChange: function() {
            $('#addInstanceProxyUrlField').toggle(this.checked);
        }
    });
    
    $('#addInstanceS3Toggle').checkbox({
        onChange: function() {
            $('#addInstanceS3Fields').toggle(this.checked);
        }
    });
    
    $('#addInstanceHmacToggle').checkbox({
        onChange: function() {
            $('#addInstanceHmacKeyField').toggle(this.checked);
            $('#addInstanceHmacKeyWarningMessage').toggle(this.checked);
        }
    });

    // Chatwoot toggle
    $('#chatwootEnabledToggle').checkbox({
        onChange: function() {
            $('#chatwootFields').toggle(this.checked);
        }
    });

    $('#chatwootSignMsgToggle').checkbox({
        onChange: function() {
            $('#chatwootSignDelimiterField').toggle(this.checked);
        }
    });
}

// Exported for other JS files (e.g. groups.js)
window.baseUrl = baseUrl;
window.apiCall = apiCall;
window.getLocalStorageItem = getLocalStorageItem;
