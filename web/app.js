(function() {
  var QUICK_REACTIONS = ["\uD83D\uDC4D", "\uD83D\uDD25", "\uD83D\uDE02", "\uD83D\uDE0E", "\uD83D\uDC80"];
  var ROOM_LABELS = {
    dm: "Diretas",
    chat: "Chats",
    media: "Midia",
    ia: "IA",
    dev: "Lab",
    vip: "VIP",
    admin: "Admin"
  };
  var ROOM_GROUP_ORDER = ["Diretas", "Chats", "Midia", "IA", "Lab", "VIP", "Admin", "Salas"];
  var PRIMARY_NAV = [
    { id: "chat-geral", slug: "chat-geral", label: "Chat Geral", copy: "Resenha aberta da base.", icon: "💬", theme: "canal base", accent: "ao vivo" },
    { id: "diretas", kind: "dm", label: "Diretas", copy: "Conversa pv sem plateia.", icon: "📨", theme: "dm", accent: "1x1" },
    { id: "fotos", slug: "fotos", label: "Fotos", copy: "Galeria viva da tropa.", icon: "🖼️", theme: "midia", accent: "prints" },
    { id: "arquivos", slug: "arquivos", label: "Arquivos", copy: "Docs, packs e drop salvo.", icon: "📦", theme: "dropzone", accent: "asset" },
    { id: "nego-dramias-ia", slug: "nego-dramias-ia", label: "Nego Dramias", copy: "IA útil, gaúcha e funcional.", icon: "🤠", theme: "assistente", accent: "ia" },
    { id: "apps-lab", slug: "apps-lab", label: "Apps Lab", copy: "Sala técnica privada do admin.", icon: "🛠️", theme: "admin lab", accent: "restrito" }
  ];
  PRIMARY_NAV.splice(5, 0, { id: "coisas", slug: "coisas", label: "Coisas", copy: "Canto estranho com senha e cheiro de segredinho.", icon: "🥔", theme: "cofre torto", accent: "senha" });
  PRIMARY_NAV.push({ id: "sus", kind: "sus", label: "???", copy: "Nao era pra clicar aqui, vivente.", icon: "👀", theme: "anomalia", accent: "sus" });
  var SUS_WARNING_LINES = [
    "Certeza que tu quer saber o que tem do outro lado?",
    "Bah... ainda da tempo de fingir que nem viu isso.",
    "Eu no teu lugar largava esse botao quieto.",
    "Ultimo aviso, vivente. Depois nao reclama do destino."
  ];
  var SUS_TARGET_URL = "https://youtu.be/L-RQo8D7c3I?si=1tlBey4A2B40GS7W";
  var SUS_POSITIONS = [
    { x: 0, y: 0 },
    { x: 28, y: -16 },
    { x: -24, y: 18 },
    { x: 18, y: 26 }
  ];
  var LOGIN_ERROR_LINES = [
    "\uD83E\uDD23 Bah... essa senha veio torta.",
    "\uD83D\uDE35 Acesso negado, campeao.",
    "\uD83D\uDE2C Tentou bonito, mas a grade segurou.",
    "\uD83E\uDD74 Senha errada. Vai sem pressa e tenta de novo.",
    "\uD83E\uDD26\u200D\u2642\uFE0F Bah tche, essa senha ta mais perdida que cusco em procissao.",
    "\uD83D\uDE02 Capaz, vivente. Tu errou e o painel deu risada.",
    "\uD83E\uDD79 Essa tentativa ai saiu mais torta que carreta em atoleiro.",
    "\uD83D\uDC80 Bah, nem o Nego Dramias assinava essa senha.",
    "\uD83D\uDE10 Volta no galpao, respira e tenta sem inventar moda.",
    "\uD83D\uDE48 Sinto muito, guri. Essa combinacao nao entra nem com reza forte.",
    "\uD83D\uDE0F Tu quase acertou. Quase igual cavalo que para no alambrado.",
    "\uD83E\uDD32 O sistema te viu chegando e trancou a porteira.",
    "\uD83D\uDC15 Mais facil o chimarrao ferver sozinho que essa senha estar certa.",
    "\uD83D\uDE36 Bah, de novo nao. Ate a matrix ficou com vergonha."
  ];
  var LOGIN_SUCCESS_LINES = [
    "\uD83D\uDE0E Acesso liberado, vivente.",
    "\u2728 Entrou liso no Painel Dief.",
    "\uD83D\uDE80 Tudo certo. Pode assumir a cabine.",
    "\uD83E\uDD18 Sistema liberado sem drama.",
    "\uD83C\uDF89 Bah tche, agora sim abriu bonito."
  ];
  var APP_DOWNLOAD_URL = "/downloads/universalD.exe";
  var UNIVERSALD_META = {
    version: "3.1.3-desktop-native",
    shortVersion: "v3.1.3",
    sizeLabel: "16.13 MB",
    sha256: "DB156B2648BAE5FB5995CFDA9655375CB959A4FFDC24DDE37F9133C242F8665F",
    changelog: [
      "Login do app desktop refeito em modo compatível com o motor nativo do exe.",
      "Sessão do painel agora tem fallback local para sobreviver melhor a reload e cookie ruim.",
      "Ícone do app regenerado da arte nova e empacotado também como .ico real.",
      "Registro nativo do protocolo universald:// pra abrir a base direto do site.",
      "Build privado renovado e pronto pra baixar com senha pela base oficial."
    ]
  };
  var SESSION_STORAGE_KEY = "painelDiefSessionId";
  var THEME_PRESETS = {
    matrix: { label: "Matrix", accent: "#7bff00" },
    obsidian: { label: "Obsidian", accent: "#90a6ff" },
    ember: { label: "Ember", accent: "#ff7a42" },
    cobalt: { label: "Cobalt", accent: "#54b8ff" },
    neon: { label: "Neon", accent: "#ff5ad4" }
  };
  var BANNER_PRESETS = {
    grid: { label: "Grid" },
    pulse: { label: "Pulse" },
    arcade: { label: "Arcade" },
    horizon: { label: "Horizon" },
    stealth: { label: "Stealth" }
  };

  var state = {
    viewer: null,
    rooms: [],
    roomAccess: {},
    online: [],
    typing: [],
    recentLogs: [],
    latestByRoom: {},
    events: [],
    polls: [],
    roomVersions: {},
    blockedUserIds: [],
    mutedUserIds: [],
    users: [],
    joinRequests: [],
    joinRequestsLoaded: false,
    pendingJoinCount: 0,
    activeRoomId: 0,
    messagesByRoom: {},
    pinnedByRoom: {},
    unread: {},
    latestUnreadByRoom: {},
    pendingAttachment: null,
    pendingUploads: [],
    replyTarget: null,
    editingMessage: null,
    selectedMember: null,
    stream: null,
    heartbeatTimer: null,
    activity: [],
    version: 0,
    previousOnlineMap: {},
    unlockRoomId: 0,
    isMobile: false,
    compactLayout: false,
    activeNavId: "chat-geral",
    filter: "all",
    roomSearch: "",
    inspectorTab: "overview",
    searchResult: null,
    searchQuery: "",
    memberSearch: "",
    memberRoleFilter: "all",
    mediaFilter: "all",
    mediaSearch: "",
    mediaSort: "recent",
    mediaPreview: null,
    highlightMessageId: 0,
    typingIdleTimer: null,
    typingSent: false,
    sendingMessage: false,
    audioEnabled: true,
    audioContext: null,
    appsNotesDraft: "",
    favoriteNavIds: [],
    focusMode: false,
    connectionStatus: "syncing",
    lastStreamAt: 0,
    appInstalled: false,
    appInstallSeenAt: 0,
    susClicks: 0,
    susPositionIndex: 0,
    utilityStripCollapsed: false,
    chatContextCollapsed: false,
    emergencyMode: false,
    desktopSidebarCollapsed: false,
    desktopInspectorCollapsed: false,
    lastToastKey: "",
    lastToastAt: 0,
    notificationCenterOpen: false
  };

  function readStoredSessionId() {
    try {
      return String(window.localStorage.getItem(SESSION_STORAGE_KEY) || "").trim();
    } catch (e) {
      return "";
    }
  }

  function storeSessionId(sessionId) {
    var clean = String(sessionId || "").trim();
    try {
      if (!clean) {
        window.localStorage.removeItem(SESSION_STORAGE_KEY);
        return;
      }
      window.localStorage.setItem(SESSION_STORAGE_KEY, clean);
    } catch (e) {}
  }

  function clearStoredSessionId() {
    storeSessionId("");
  }

  function q(id) {
    return document.getElementById(id);
  }

  function syncViewportMetrics() {
    var root = document.documentElement;
    var viewport = window.visualViewport || null;
    var width = viewport && viewport.width ? viewport.width : window.innerWidth;
    var height = viewport && viewport.height ? viewport.height : window.innerHeight;
    var keyboardOffset = Math.max(0, Math.round(window.innerHeight - height));
    if (!root) {
      return;
    }
    root.style.setProperty("--viewport-width", Math.max(320, Math.round(width)) + "px");
    root.style.setProperty("--viewport-height", Math.max(320, Math.round(height)) + "px");
    root.style.setProperty("--keyboard-offset", keyboardOffset + "px");
    document.body.classList.toggle("keyboard-open", keyboardOffset > 120);
  }

  var layoutRefreshRaf = 0;

  function scheduleLayoutRefresh() {
    if (layoutRefreshRaf) {
      return;
    }
    layoutRefreshRaf = window.requestAnimationFrame(function() {
      layoutRefreshRaf = 0;
      syncViewportMetrics();
      detectMobile();
    });
  }

  function safeScrollIntoView(target, options) {
    if (!target || typeof target.scrollIntoView !== "function") {
      return;
    }
    try {
      if (options) {
        target.scrollIntoView(options);
      } else {
        target.scrollIntoView();
      }
      return;
    } catch (err) {
    }
    try {
      target.scrollIntoView();
    } catch (ignored) {
    }
  }

  function enterEmergencyMode(message) {
    var loginError = q("login-error");
    state.emergencyMode = true;
    document.body.classList.add("emergency-mode", "login-mode");
    document.body.classList.remove(
      "panel-mode",
      "modal-open",
      "drawer-open",
      "sidebar-open",
      "inspector-open",
      "utility-strip-collapsed",
      "chat-context-collapsed",
      "focus-mode"
    );
    if (q("panel-view")) {
      q("panel-view").classList.add("hidden");
    }
    if (q("login-view")) {
      q("login-view").classList.remove("hidden");
      q("login-view").scrollTop = 0;
    }
    if (loginError) {
      loginError.classList.remove("hidden");
      loginError.textContent = message || "Modo de compatibilidade ativado. Tenta entrar de novo.";
    }
    try {
      window.scrollTo(0, 0);
    } catch (e) {}
  }

  function clearPanelEmergencyLayout() {
    var panelView = q("panel-view");
    var panelShell = q("panel-shell");
    var workspace = q("workspace");
    var workspaceGrid = q("workspace-grid");
    var channelPanel = q("channel-panel");
    var conversation = q("conversation-panel");
    var inspector = q("inspector-panel");
    document.body.classList.remove("panel-emergency");
    [panelView, panelShell, workspace, workspaceGrid, channelPanel, conversation, inspector].forEach(function(node) {
      if (!node || !node.style) {
        return;
      }
      node.style.display = "";
      node.style.height = "";
      node.style.minHeight = "";
      node.style.maxHeight = "";
      node.style.overflow = "";
      node.style.gridTemplateColumns = "";
      node.style.gridTemplateRows = "";
      node.style.padding = "";
      node.style.visibility = "";
      node.style.opacity = "";
      node.style.pointerEvents = "";
      node.style.transform = "";
      node.style.position = "";
      node.style.inset = "";
      node.style.width = "";
    });
  }

  function activatePanelEmergencyLayout() {
    var panelView = q("panel-view");
    var panelShell = q("panel-shell");
    var workspace = q("workspace");
    var workspaceGrid = q("workspace-grid");
    var conversation = q("conversation-panel");
    var channelPanel = q("channel-panel");
    var inspector = q("inspector-panel");
    var topbar = q("topbar");
    state.emergencyMode = true;
    document.body.classList.add("panel-emergency", "panel-mode");
    document.body.classList.remove(
      "login-mode",
      "focus-mode",
      "modal-open",
      "drawer-open",
      "sidebar-open",
      "inspector-open",
      "sidebar-collapsed",
      "inspector-collapsed",
      "utility-strip-collapsed",
      "chat-context-collapsed"
    );
    if (q("login-view")) {
      q("login-view").classList.add("hidden");
    }
    if (panelView) {
      panelView.classList.remove("hidden");
      panelView.style.display = "block";
      panelView.style.overflow = "auto";
      panelView.style.padding = "1rem";
      panelView.scrollTop = 0;
    }
    if (panelShell) {
      panelShell.style.display = "grid";
      panelShell.style.gridTemplateColumns = "minmax(0, 1fr)";
      panelShell.style.height = "auto";
      panelShell.style.minHeight = "calc(var(--viewport-height) - 2rem)";
      panelShell.style.overflow = "visible";
    }
    if (workspace) {
      workspace.style.display = "grid";
      workspace.style.gridTemplateRows = "auto minmax(0, 1fr)";
      workspace.style.height = "auto";
      workspace.style.minHeight = "calc(var(--viewport-height) - 2.4rem)";
      workspace.style.overflow = "visible";
    }
    if (workspaceGrid) {
      workspaceGrid.style.display = "grid";
      workspaceGrid.style.gridTemplateColumns = "minmax(0, 1fr)";
      workspaceGrid.style.height = "auto";
      workspaceGrid.style.minHeight = "0";
      workspaceGrid.style.overflow = "visible";
    }
    if (conversation) {
      conversation.style.display = "grid";
      conversation.style.gridTemplateRows = "auto auto auto auto auto minmax(260px, 1fr) auto auto";
      conversation.style.height = "auto";
      conversation.style.minHeight = "calc(var(--viewport-height) - 8rem)";
      conversation.style.maxHeight = "none";
      conversation.style.overflow = "hidden";
    }
    if (channelPanel) {
      channelPanel.style.display = "none";
      channelPanel.style.visibility = "hidden";
      channelPanel.style.opacity = "0";
      channelPanel.style.pointerEvents = "none";
      channelPanel.style.transform = "none";
    }
    if (inspector) {
      inspector.style.display = "none";
      inspector.style.visibility = "hidden";
      inspector.style.opacity = "0";
      inspector.style.pointerEvents = "none";
      inspector.style.transform = "none";
    }
    if (topbar) {
      try {
        safeScrollIntoView(topbar, { block: "start", behavior: "auto" });
      } catch (e) {}
    }
  }

  function ensureFieldVisible(target) {
    if (!target || !(state.compactLayout || state.isMobile)) {
      return;
    }
    window.setTimeout(function() {
      var scroller = null;
      var targetRect = null;
      var viewportHeight = 0;
      var keyboardAwareBottom = 0;
      var outsideViewport = false;
      try {
        scroller = target.closest && target.closest(".modal-card, .login-card, .sidebar-scroll, .inspector-body, #message-stream, #panel-view, #login-view");
      } catch (closestErr) {}
      try {
        targetRect = target.getBoundingClientRect();
        viewportHeight = window.visualViewport && window.visualViewport.height ? window.visualViewport.height : window.innerHeight;
        keyboardAwareBottom = Math.max(140, viewportHeight - 24);
        outsideViewport = !!(targetRect && (targetRect.top < 14 || targetRect.bottom > keyboardAwareBottom));
      } catch (rectErr) {}
      if (scroller && scroller !== document.body && scroller !== document.documentElement) {
        try {
          var scrollerRect = scroller.getBoundingClientRect();
          var delta = 0;
          if (targetRect.bottom > scrollerRect.bottom - 18) {
            delta = targetRect.bottom - (scrollerRect.bottom - 18);
          } else if (targetRect.top < scrollerRect.top + 18) {
            delta = targetRect.top - (scrollerRect.top + 18);
          }
          if (delta) {
            scroller.scrollTop += delta;
          }
        } catch (scrollErr) {}
        return;
      }
      if (outsideViewport) {
        safeScrollIntoView(target, {
          block: state.isMobile ? "nearest" : "center",
          inline: "nearest",
          behavior: state.isMobile ? "auto" : "smooth"
        });
      }
    }, 120);
  }

  function isPanelSessionManagedPath(path) {
    var clean = String(path || "").trim();
    return /^\/api\/panel\//.test(clean) && clean !== "/api/panel/login";
  }

  function esc(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function safeAvatarUrl(raw) {
    var value = String(raw || "").trim();
    var parsed;
    if (!value) {
      return "";
    }
    if (/^\/uploads\/[A-Za-z0-9._\-]+$/i.test(value)) {
      return value;
    }
    try {
      parsed = new URL(value, window.location.origin);
    } catch (e) {
      return "";
    }
    if (parsed.origin === window.location.origin && /^\/uploads\/[A-Za-z0-9._\-]+$/i.test(parsed.pathname)) {
      return parsed.pathname + parsed.search + parsed.hash;
    }
    if (parsed.protocol === "https:") {
      return parsed.toString();
    }
    return "";
  }

  function safeBannerUrl(raw) {
    var value = String(raw || "").trim();
    var parsed;
    if (!value) {
      return "";
    }
    if (/^\/uploads\/[A-Za-z0-9._\-]+$/i.test(value)) {
      return value;
    }
    try {
      parsed = new URL(value, window.location.origin);
    } catch (e) {
      return "";
    }
    if (parsed.origin === window.location.origin && /^\/uploads\/[A-Za-z0-9._\-]+$/i.test(parsed.pathname)) {
      return parsed.pathname + parsed.search + parsed.hash;
    }
    if (parsed.protocol === "https:") {
      return parsed.toString();
    }
    return "";
  }

  function cssUrl(value) {
    return encodeURI(String(value || ""))
      .replace(/'/g, "%27")
      .replace(/\(/g, "%28")
      .replace(/\)/g, "%29");
  }

  function safeAttachmentUrl(raw) {
    var value = String(raw || "").trim();
    var parsed;
    if (!value) {
      return "";
    }
    if (/^\/uploads\/[A-Za-z0-9._\-]+$/i.test(value)) {
      return value;
    }
    try {
      parsed = new URL(value, window.location.origin);
    } catch (e) {
      return "";
    }
    if (parsed.origin === window.location.origin && /^\/uploads\/[A-Za-z0-9._\-]+$/i.test(parsed.pathname)) {
      return parsed.pathname + parsed.search + parsed.hash;
    }
    return "";
  }

  function initials(name) {
    var clean = String(name || "").trim();
    if (!clean) {
      return "PD";
    }
    var parts = clean.split(/\s+/);
    if (parts.length === 1) {
      return parts[0].slice(0, 2).toUpperCase();
    }
    return (parts[0].slice(0, 1) + parts[1].slice(0, 1)).toUpperCase();
  }

  function roomById(id) {
    var i;
    for (i = 0; i < state.rooms.length; i++) {
      if (Number(state.rooms[i].id) === Number(id)) {
        return state.rooms[i];
      }
    }
    return null;
  }

  function roomBySlug(slug) {
    var i;
    for (i = 0; i < state.rooms.length; i++) {
      if (String(state.rooms[i].slug || "") === String(slug || "")) {
        return state.rooms[i];
      }
    }
    return null;
  }

  function activeRoom() {
    return roomById(state.activeRoomId);
  }

  function isAppsLabRoom(room) {
    return !!(room && room.slug === "apps-lab");
  }

  function isSusView() {
    return state.activeNavId === "sus";
  }

  function currentSusWarning() {
    var index = Math.max(0, Math.min(SUS_WARNING_LINES.length - 1, Number(state.susClicks || 0)));
    return SUS_WARNING_LINES[index] || SUS_WARNING_LINES[0];
  }

  function currentSusPosition() {
    var index = Math.max(0, Math.min(SUS_POSITIONS.length - 1, Number(state.susPositionIndex || 0)));
    return SUS_POSITIONS[index] || SUS_POSITIONS[0];
  }

  function presenceByUserId(userId) {
    var i;
    for (i = 0; i < state.online.length; i++) {
      if (Number(state.online[i].userId) === Number(userId)) {
        return state.online[i];
      }
    }
    return null;
  }

  function isBlockedUser(userId) {
    return state.blockedUserIds.indexOf(Number(userId)) >= 0;
  }

  function isMutedUser(userId) {
    return state.mutedUserIds.indexOf(Number(userId)) >= 0;
  }

  function accessForRoom(room) {
    if (!room) {
      return "open";
    }
    return state.roomAccess[room.slug] || "open";
  }

  function onlineMap(items) {
    var map = {};
    var i;
    for (i = 0; i < items.length; i++) {
      map[String(items[i].userId)] = !!items[i].online;
    }
    return map;
  }

  function formatTime(raw) {
    try {
      return new Date(raw).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" });
    } catch (e) {
      return "--:--";
    }
  }

  function formatDateTime(raw) {
    try {
      return new Date(raw).toLocaleString("pt-BR");
    } catch (e) {
      return "--";
    }
  }

  function formatRelative(raw) {
    try {
      var delta = new Date(raw).getTime() - Date.now();
      var minutes = Math.round(delta / 60000);
      var absMinutes = Math.abs(minutes);
      if (absMinutes < 1) {
        return "agora";
      }
      if (absMinutes < 60) {
        return minutes > 0 ? ("em " + absMinutes + " min") : (absMinutes + " min atras");
      }
      var hours = Math.round(absMinutes / 60);
      if (hours < 24) {
        return minutes > 0 ? ("em " + hours + " h") : (hours + " h atras");
      }
      var days = Math.round(hours / 24);
      return minutes > 0 ? ("em " + days + " d") : (days + " d atras");
    } catch (e) {
      return "--";
    }
  }

  function bytesLabel(size) {
    var units = ["B", "KB", "MB", "GB"];
    var value = Number(size || 0);
    var index = 0;
    while (value >= 1024 && index < units.length - 1) {
      value = value / 1024;
      index++;
    }
    return value.toFixed(index === 0 ? 0 : 1) + " " + units[index];
  }

  function totalAttachmentBytes(items) {
    return (items || []).reduce(function(sum, item) {
      return sum + Number(item && item.attachment && item.attachment.sizeBytes || 0);
    }, 0);
  }

  function pickRandom(items) {
    if (!items || !items.length) {
      return "";
    }
    return items[Math.floor(Math.random() * items.length)];
  }

  function themeMeta(theme) {
    return THEME_PRESETS[String(theme || "matrix").toLowerCase()] || THEME_PRESETS.matrix;
  }

  function themeLabel(theme) {
    return themeMeta(theme).label;
  }

  function themeAccent(theme) {
    return themeMeta(theme).accent;
  }

  function bannerMeta(bannerPreset) {
    return BANNER_PRESETS[String(bannerPreset || "grid").toLowerCase()] || BANNER_PRESETS.grid;
  }

  function bannerLabel(bannerPreset) {
    return bannerMeta(bannerPreset).label;
  }

  function normalizeAccentColor(value, fallback) {
    var clean = String(value || "").trim();
    if (/^#[0-9a-f]{3}$/i.test(clean)) {
      return "#" + clean.slice(1).split("").map(function(char) { return char + char; }).join("");
    }
    if (/^#[0-9a-f]{6}$/i.test(clean) || /^#[0-9a-f]{8}$/i.test(clean)) {
      return clean;
    }
    return String(fallback || "#7bff00");
  }

  function panelBannerStyle(theme, accent, bannerPreset, bannerUrl) {
    var base = normalizeAccentColor(themeAccent(theme), "#7bff00");
    var tone = normalizeAccentColor(accent, base);
    var cleanBannerUrl = safeBannerUrl(bannerUrl);
    var bannerLayer = cleanBannerUrl
      ? "linear-gradient(180deg, rgba(6,10,18,0.18), rgba(6,10,18,0.64)), url('" + cssUrl(cleanBannerUrl) + "') center/cover no-repeat, "
      : "";
    switch (String(bannerPreset || "grid").toLowerCase()) {
      case "pulse":
        return bannerLayer + "radial-gradient(circle at 18% 24%, " + tone + "88 0%, transparent 24%), " +
          "radial-gradient(circle at 78% 22%, " + base + "44 0%, transparent 26%), " +
          "radial-gradient(circle at 50% 110%, " + tone + "55 0%, transparent 42%), " +
          "linear-gradient(135deg, rgba(6,10,18,0.98) 0%, rgba(7,22,20,0.94) 48%, rgba(8,12,20,0.98) 100%)";
      case "arcade":
        return bannerLayer + "linear-gradient(120deg, rgba(255,255,255,0.04) 0 9%, transparent 9% 18%, rgba(255,255,255,0.03) 18% 26%, transparent 26% 36%, rgba(255,255,255,0.02) 36% 100%), " +
          "linear-gradient(135deg, " + tone + "22 0%, transparent 40%), " +
          "linear-gradient(90deg, rgba(7,13,20,0.98), rgba(11,18,29,0.95) 45%, rgba(8,12,20,0.98))";
      case "horizon":
        return bannerLayer + "radial-gradient(circle at 50% 130%, " + tone + "55 0%, transparent 34%), " +
          "linear-gradient(180deg, " + base + "22 0%, transparent 24%), " +
          "linear-gradient(135deg, rgba(5,10,18,0.98) 0%, rgba(9,17,30,0.94) 52%, rgba(8,14,22,0.98) 100%)";
      case "stealth":
        return bannerLayer + "linear-gradient(90deg, rgba(255,255,255,0.05) 0 2%, transparent 2% 14%, rgba(255,255,255,0.02) 14% 16%, transparent 16% 100%), " +
          "radial-gradient(circle at 85% 14%, " + tone + "33 0%, transparent 22%), " +
          "linear-gradient(135deg, rgba(4,8,12,0.99), rgba(8,12,18,0.96) 55%, rgba(6,8,12,0.99))";
      default:
        return bannerLayer + "radial-gradient(circle at 14% 18%, " + tone + "66 0%, transparent 34%), " +
          "radial-gradient(circle at 84% 12%, " + base + "55 0%, transparent 28%), " +
          "linear-gradient(135deg, rgba(6,10,18,0.98) 0%, rgba(10,18,28,0.94) 48%, rgba(8,12,20,0.98) 100%)";
    }
  }

  function normalizeText(value) {
    return String(value || "")
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .trim();
  }

  function filterLabel(value) {
    if (value === "chat") {
      return "chat";
    }
    if (value === "direct") {
      return "diretas";
    }
    if (value === "media") {
      return "midia";
    }
    if (value === "dev") {
      return "lab";
    }
    if (value === "secure") {
      return "seguras";
    }
    return "todas";
  }

  function dmPeerName(room) {
    var raw = String(room && room.name || "").trim();
    if (!raw) {
      return "Conversa direta";
    }
    var parts = raw.split(/\s+x\s+/i).map(function(item) {
      return item.trim();
    }).filter(Boolean);
    if (parts.length < 2) {
      return raw;
    }
    var viewerKeys = {};
    [state.viewer && state.viewer.displayName, state.viewer && state.viewer.username].forEach(function(name) {
      var key = normalizeText(name);
      if (key) {
        viewerKeys[key] = true;
      }
    });
    var peers = parts.filter(function(name) {
      return !viewerKeys[normalizeText(name)];
    });
    return peers[0] || parts[0];
  }

  function displayRoomName(room) {
    if (!room) {
      return "Sala";
    }
    if (room.scope === "dm") {
      return dmPeerName(room);
    }
    return room.name || "Sala";
  }

  function displayRoomDescription(room) {
    if (!room) {
      return "Sem descricao.";
    }
    if (room.scope === "dm") {
      return "Conversa direta, privada e persistente.";
    }
    if (room.category === "media") {
      return room.description || "Espaco para fotos, videos, audios e arquivos.";
    }
    if (room.category === "ia") {
      return room.description || "Fala com o Nego Dramias e pede ajuda no sistema.";
    }
    if (room.category === "dev") {
      return room.description || "Area tecnica para apps, comandos, snippets e testes.";
    }
    if (room.category === "admin") {
      return room.description || "Sala reservada para controle, moderacao e logs.";
    }
    if (room.vipOnly || room.passwordProtected) {
      return room.description || "Canal protegido por senha, cargo ou permissao.";
    }
    return room.description || "Canal principal do painel.";
  }

  function displayRoomIcon(room) {
    if (!room) {
      return "RM";
    }
    if (room.scope === "dm") {
      return initials(dmPeerName(room));
    }
    return room.icon || "RM";
  }

  function primaryDirectRoom() {
    return directRoomsSorted()[0] || null;
  }

  function directRoomsSorted() {
    var dms = state.rooms.filter(function(room) {
      return room.scope === "dm";
    }).slice(0);
    dms.sort(function(a, b) {
      var stampA = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0;
      var stampB = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0;
      return stampB - stampA;
    });
    return dms;
  }

  function navDefinition(id) {
    var i;
    for (i = 0; i < PRIMARY_NAV.length; i++) {
      if (PRIMARY_NAV[i].id === id) {
        return PRIMARY_NAV[i];
      }
    }
    return PRIMARY_NAV[0];
  }

  function hasPrimaryNavId(id) {
    var i;
    for (i = 0; i < PRIMARY_NAV.length; i++) {
      if (PRIMARY_NAV[i].id === id) {
        return true;
      }
    }
    return false;
  }

  function navRoom(nav) {
    if (!nav) {
      return null;
    }
    if (nav.kind === "sus") {
      return null;
    }
    if (nav.kind === "dm") {
      return primaryDirectRoom();
    }
    if (nav.slug) {
      return roomBySlug(nav.slug);
    }
    return null;
  }

  function navIdForRoom(room) {
    if (!room) {
      return "chat-geral";
    }
    if (room.scope === "dm") {
      return "diretas";
    }
    return room.slug || "chat-geral";
  }

  function isMembersHub() {
    return state.activeNavId === "members";
  }

  function isDirectHubEmpty() {
    return state.activeNavId === "diretas" && !primaryDirectRoom();
  }

  function isHubView() {
    return isMembersHub() || isDirectHubEmpty();
  }

  function navVisible(nav) {
    if (!nav) {
      return false;
    }
    if (nav.kind === "sus") {
      return true;
    }
    if (nav.kind === "dm") {
      return true;
    }
    return !!navRoom(nav);
  }

  function navSearchAllows(nav) {
    var room = navRoom(nav);
    var query = normalizeText(state.roomSearch);
    var haystack = [
      nav.label,
      nav.copy,
      room && room.name,
      room && room.description,
      room && room.lastMessagePreview,
      room && roomKindLabel(room)
    ].map(normalizeText).join(" ");
    if (!query) {
      return true;
    }
    return haystack.indexOf(query) >= 0;
  }

  function sortNavItems(items) {
    return (items || []).slice(0).sort(function(a, b) {
      var aFav = isFavoriteNavId(a.id) ? 0 : 1;
      var bFav = isFavoriteNavId(b.id) ? 0 : 1;
      if (aFav !== bFav) {
        return aFav - bFav;
      }
      return PRIMARY_NAV.findIndex(function(item) { return item.id === a.id; }) - PRIMARY_NAV.findIndex(function(item) { return item.id === b.id; });
    });
  }

  function composerPlaceholder(room) {
    if (!room) {
      return "Escolhe uma sala e manda a letra...";
    }
    if (room.scope === "dm") {
      return "Escreve direto pra " + displayRoomName(room) + " sem plateia...";
    }
    if (room.category === "media") {
      return "Solta foto, video, audio ou arquivo com contexto...";
    }
    if (room.category === "ia") {
      return "Pergunta pro Nego Dramias e manda o drama...";
    }
    if (room.category === "dev") {
      return "Manda codigo, comando, log ou ideia tecnica...";
    }
    if (room.vipOnly || room.passwordProtected) {
      return "Papo reservado. Nada de vacilo por aqui...";
    }
    return "Manda mensagem, usa /help, responde alguem ou sobe um anexo...";
  }

  function roomKind(room) {
    if (!room) {
      return "chat";
    }
    if (room.scope === "dm") {
      return "dm";
    }
    if (room.category === "media") {
      return "media";
    }
    if (room.category === "dev" || room.category === "ia") {
      return "dev";
    }
    if (room.category === "admin") {
      return "admin";
    }
    if (room.vipOnly || room.passwordProtected) {
      return "secure";
    }
    return "chat";
  }

  function roomKindLabel(room) {
    var kind = roomKind(room);
    if (kind === "dm") {
      return "dm";
    }
    if (kind === "media") {
      return "midia";
    }
    if (kind === "dev") {
      return room.category === "ia" ? "ia" : "lab";
    }
    if (kind === "admin") {
      return "admin";
    }
    if (kind === "secure") {
      return "segura";
    }
    return "canal";
  }

  function groupRoomLabel(room) {
    if (room.scope === "dm") {
      return "Diretas";
    }
    return ROOM_LABELS[room.category] || "Salas";
  }

  function filterAllowsRoom(room) {
    var access = accessForRoom(room);
    if (state.filter === "all") {
      return true;
    }
    if (state.filter === "chat") {
      return room.category === "chat" || room.category === "ia";
    }
    if (state.filter === "direct") {
      return room.scope === "dm";
    }
    if (state.filter === "media") {
      return room.category === "media";
    }
    if (state.filter === "dev") {
      return room.category === "dev" || room.category === "ia";
    }
    if (state.filter === "secure") {
      return access === "locked" || access === "vip" || access === "admin";
    }
    return true;
  }

  function activeRoomMembers() {
    var roomId = Number(state.activeRoomId || 0);
    return state.online.filter(function(item) {
      return item.online && Number(item.roomId) === roomId;
    });
  }

  function currentRoomMessages() {
    return state.messagesByRoom[String(state.activeRoomId)] || [];
  }

  function currentRoomAttachments() {
    return currentRoomMessages().filter(function(message) {
      return !!message.attachment && !message.blockedByViewer;
    });
  }

  function roomAttachmentsForPreview(roomId) {
    var messages = state.messagesByRoom[String(roomId)] || [];
    return messages.filter(function(message) {
      return !!message.attachment && !message.blockedByViewer;
    }).sort(compareAttachmentItems);
  }

  function pendingUploads() {
    return state.pendingUploads || [];
  }

  function readyPendingUploads() {
    return pendingUploads().filter(function(item) {
      return item && item.status === "ready" && item.attachment;
    });
  }

  function hasUploadingPendingUploads() {
    return pendingUploads().some(function(item) {
      return item && item.status === "uploading";
    });
  }

  function hasErroredPendingUploads() {
    return pendingUploads().some(function(item) {
      return item && item.status === "error";
    });
  }

  function syncPendingAttachmentAlias() {
    state.pendingAttachment = readyPendingUploads().length ? readyPendingUploads()[0].attachment : null;
  }

  function mediaMatchesFilters(message) {
    var attachment = message && message.attachment;
    var query;
    var haystack;
    if (!attachment) {
      return false;
    }
    if (state.mediaFilter !== "all" && attachment.kind !== state.mediaFilter) {
      return false;
    }
    query = normalizeText(state.mediaSearch);
    if (!query) {
      return true;
    }
    haystack = [
      attachment.name,
      attachment.kind,
      attachment.contentType,
      attachment.extension,
      message.authorName,
      message.body
    ].map(normalizeText).join(" ");
    return haystack.indexOf(query) >= 0;
  }

  function filteredRoomAttachments() {
    return currentRoomAttachments().filter(mediaMatchesFilters).sort(compareAttachmentItems);
  }

  function compareAttachmentItems(a, b) {
    var mode = state.mediaSort || "recent";
    var timeA = new Date(a.createdAt).getTime();
    var timeB = new Date(b.createdAt).getTime();
    var nameA = String(a.attachment && a.attachment.name || "");
    var nameB = String(b.attachment && b.attachment.name || "");
    var sizeA = Number(a.attachment && a.attachment.sizeBytes || 0);
    var sizeB = Number(b.attachment && b.attachment.sizeBytes || 0);
    if (mode === "oldest") {
      return timeA - timeB || nameA.localeCompare(nameB, "pt-BR");
    }
    if (mode === "name") {
      return nameA.localeCompare(nameB, "pt-BR") || (timeB - timeA);
    }
    if (mode === "size") {
      return sizeB - sizeA || (timeB - timeA);
    }
    return timeB - timeA || nameA.localeCompare(nameB, "pt-BR");
  }

  function filteredPresenceItems() {
    var query = normalizeText(state.memberSearch);
    return state.online.slice(0).sort(function(a, b) {
      if (a.online === b.online) {
        return String(a.displayName || a.username || "").localeCompare(String(b.displayName || b.username || ""), "pt-BR");
      }
      return a.online ? -1 : 1;
    }).filter(function(person) {
      var roleFilter = String(state.memberRoleFilter || "all");
      var haystack;
      if (roleFilter === "online" && !person.online) {
        return false;
      }
      if (roleFilter !== "all" && roleFilter !== "online" && String(person.role || "member") !== roleFilter) {
        return false;
      }
      if (!query) {
        return true;
      }
      haystack = [
        person.displayName,
        person.username,
        person.role,
        person.status,
        person.statusText,
        person.bio
      ].map(normalizeText).join(" ");
      return haystack.indexOf(query) >= 0;
    });
  }

  function memberLocalStats(userId) {
    var stats = {
      messageCount: 0,
      attachmentCount: 0,
      roomMap: {},
      lastMessage: null,
      lastRoom: null
    };
    Object.keys(state.messagesByRoom || {}).forEach(function(roomId) {
      (state.messagesByRoom[roomId] || []).forEach(function(message) {
        var authorId = Number(message.authorId || message.userId || 0);
        if (authorId !== Number(userId)) {
          return;
        }
        stats.messageCount += 1;
        stats.roomMap[String(message.roomId)] = true;
        if (message.attachment) {
          stats.attachmentCount += 1;
        }
        if (!stats.lastMessage || new Date(message.createdAt).getTime() > new Date(stats.lastMessage.createdAt).getTime()) {
          stats.lastMessage = message;
          stats.lastRoom = roomById(message.roomId);
        }
      });
    });
    stats.roomCount = Object.keys(stats.roomMap).length;
    return stats;
  }

  function currentViewerLocalStats() {
    if (!state.viewer) {
      return {
        messageCount: 0,
        attachmentCount: 0,
        roomCount: 0,
        lastMessage: null,
        lastRoom: null
      };
    }
    return memberLocalStats(state.viewer.id);
  }

  function unreadTotalCount() {
    return Object.keys(state.unread || {}).reduce(function(sum, roomId) {
      return sum + Number(state.unread[roomId] || 0);
    }, 0);
  }

  function sharedVisibleRoomsForUser(userId) {
    var map = {};
    var list = [];
    Object.keys(state.messagesByRoom || {}).forEach(function(roomId) {
      var room = roomById(roomId);
      var hasViewer = false;
      var hasMember = false;
      if (!room) {
        return;
      }
      (state.messagesByRoom[roomId] || []).forEach(function(message) {
        var authorId = Number(message.authorId || message.userId || 0);
        if (authorId === Number(state.viewer && state.viewer.id)) {
          hasViewer = true;
        }
        if (authorId === Number(userId)) {
          hasMember = true;
        }
      });
      if (hasViewer && hasMember && !map[String(room.id)]) {
        map[String(room.id)] = true;
        list.push(room);
      }
    });
    return list;
  }

  function dashboardQuicklineItems(room, mode, attachments) {
    var onlineNow = state.online.filter(function(item) { return item.online; });
    var chips = [];
    if (isSusView()) {
      chips.push("aba suspeita armada");
      chips.push((4 - Math.min(4, Number(state.susClicks || 0))) + " avisos antes do abismo");
      chips.push("nao era pra clicar");
      return chips;
    }
    if (isMembersHub()) {
      chips.push(onlineNow.length + " online agora");
      chips.push(state.online.filter(function(item) { return String(item.role || "") === "admin" && item.online; }).length + " admins ativos");
      chips.push(state.online.filter(function(item) { return String(item.role || "") === "vip" && item.online; }).length + " vip na base");
      if (state.memberSearch || state.memberRoleFilter !== "all") {
        chips.push("filtro de membros ligado");
      }
      return chips;
    }
    if (isDirectHubEmpty()) {
      chips.push("abre um membro pra puxar DM");
      chips.push(onlineNow.length + " contatos online");
      chips.push(state.rooms.filter(function(item) { return item.scope === "dm"; }).length + " dms existentes");
      return chips;
    }
    if (!room) {
      return chips;
    }
    if (isAppsLabRoom(room)) {
      chips.push("build " + UNIVERSALD_META.shortVersion);
      chips.push(UNIVERSALD_META.sizeLabel);
      chips.push(activeRoomMembers().length + " admin na sala");
      chips.push(state.appInstalled ? "app ja visto nesse device" : "app ainda nao visto");
    }
    chips.push(currentRoomMessages().length + " msgs visiveis");
    chips.push(attachments.length + " anexos");
    chips.push(activeRoomMembers().length + " na sala");
    chips.push(Number(state.unread[String(state.activeRoomId)] || 0) + " nao lidas");
    chips.push(isFavoriteNavId(state.activeNavId) ? "area fixada" : "area solta");
    if (room.lastMessageAt) {
      chips.push("ultimo pulso " + formatRelative(room.lastMessageAt));
    }
    chips.push(roomModeLabel(room, mode) + " // " + roomPulseLabel(room));
    return chips;
  }

  function mediaSortLabel(value) {
    if (value === "oldest") {
      return "mais antigos";
    }
    if (value === "name") {
      return "nome";
    }
    if (value === "size") {
      return "tamanho";
    }
    return "mais recentes";
  }

  function uploadLimitLabel() {
    if (!state.viewer) {
      return "30MB";
    }
    if (state.viewer.role === "owner" || state.viewer.role === "admin") {
      return "120MB";
    }
    if (state.viewer.role === "vip") {
      return "60MB";
    }
    return "30MB";
  }

  function guideStorageKey() {
    var viewerPart = state.viewer && state.viewer.id ? String(state.viewer.id) : "guest";
    return "painel-dief-guide-v1.4.4-" + viewerPart;
  }

  function hasSeenGuide() {
    try {
      return window.localStorage.getItem(guideStorageKey()) === "1";
    } catch (e) {
      return false;
    }
  }

  function markGuideSeen() {
    try {
      window.localStorage.setItem(guideStorageKey(), "1");
    } catch (e) {}
  }

  function clearGuideSeen() {
    try {
      window.localStorage.removeItem(guideStorageKey());
    } catch (e) {}
  }

  function personStatusCopy(person) {
    var text = String(person && person.statusText || "").trim();
    if (!text) {
      return "";
    }
    return text;
  }

  function parsePollOptions(raw) {
    return String(raw || "")
      .split(/\r?\n/)
      .map(function(item) { return String(item || "").trim(); })
      .filter(Boolean);
  }

  function notesStorageKey() {
    return "painel-dief.apps-notes." + Number(state.viewer && state.viewer.id || 0);
  }

  function favoriteNavStorageKey() {
    return "painel-dief.favorite-navs." + Number(state.viewer && state.viewer.id || 0);
  }

  function focusModeStorageKey() {
    return "painel-dief.focus-mode." + Number(state.viewer && state.viewer.id || 0);
  }

  function chatContextStorageKey() {
    return "painel-dief.chat-context." + Number(state.viewer && state.viewer.id || 0);
  }

  function utilityStripStorageKey() {
    return "painel-dief.utility-strip." + Number(state.viewer && state.viewer.id || 0);
  }

  function appInstallStorageKey() {
    return "painel-dief.app-install-state";
  }

  function roomDraftStorageKey(roomId) {
    return "painel-dief.room-draft." + Number(state.viewer && state.viewer.id || 0) + "." + Number(roomId || 0);
  }

  function activeRoomStorageKey() {
    return "painel-dief.active-room." + Number(state.viewer && state.viewer.id || 0);
  }

  function activeNavStorageKey() {
    return "painel-dief.active-nav." + Number(state.viewer && state.viewer.id || 0);
  }

  function loadAppsNotes() {
    try {
      return window.localStorage.getItem(notesStorageKey()) || "";
    } catch (err) {
      return "";
    }
  }

  function saveAppsNotes(value) {
    try {
      window.localStorage.setItem(notesStorageKey(), String(value || ""));
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadFavoriteNavIds() {
    try {
      var raw = window.localStorage.getItem(favoriteNavStorageKey()) || "[]";
      var items = JSON.parse(raw);
      if (!Array.isArray(items)) {
        return [];
      }
      return items.map(String);
    } catch (err) {
      return [];
    }
  }

  function saveFavoriteNavIds() {
    try {
      window.localStorage.setItem(favoriteNavStorageKey(), JSON.stringify(state.favoriteNavIds || []));
      return true;
    } catch (err) {
      return false;
    }
  }

  function isFavoriteNavId(navId) {
    return (state.favoriteNavIds || []).indexOf(String(navId)) >= 0;
  }

  function loadFocusMode() {
    try {
      return window.localStorage.getItem(focusModeStorageKey()) === "1";
    } catch (err) {
      return false;
    }
  }

  function saveFocusMode() {
    try {
      window.localStorage.setItem(focusModeStorageKey(), state.focusMode ? "1" : "0");
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadChatContextCollapsed() {
    try {
      var raw = window.localStorage.getItem(chatContextStorageKey());
      if (raw === null) {
        return !!state.compactLayout;
      }
      return raw === "1";
    } catch (err) {
      return !!state.compactLayout;
    }
  }

  function saveChatContextCollapsed() {
    try {
      window.localStorage.setItem(chatContextStorageKey(), state.chatContextCollapsed ? "1" : "0");
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadUtilityStripCollapsed() {
    try {
      var raw = window.localStorage.getItem(utilityStripStorageKey());
      if (raw === null) {
        return !!state.compactLayout;
      }
      return raw === "1";
    } catch (err) {
      return !!state.compactLayout;
    }
  }

  function saveUtilityStripCollapsed() {
    try {
      window.localStorage.setItem(utilityStripStorageKey(), state.utilityStripCollapsed ? "1" : "0");
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadAppInstallState() {
    try {
      var raw = window.localStorage.getItem(appInstallStorageKey()) || "";
      var parsed = raw ? JSON.parse(raw) : {};
      state.appInstalled = !!parsed.installed;
      state.appInstallSeenAt = Number(parsed.seenAt || 0);
    } catch (err) {
      state.appInstalled = false;
      state.appInstallSeenAt = 0;
    }
  }

  function saveAppInstallState(installed) {
    state.appInstalled = !!installed;
    state.appInstallSeenAt = installed ? Date.now() : 0;
    try {
      window.localStorage.setItem(appInstallStorageKey(), JSON.stringify({
        installed: !!installed,
        seenAt: state.appInstallSeenAt
      }));
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadRoomDraft(roomId) {
    try {
      return window.localStorage.getItem(roomDraftStorageKey(roomId)) || "";
    } catch (err) {
      return "";
    }
  }

  function saveRoomDraft(roomId, value) {
    try {
      window.localStorage.setItem(roomDraftStorageKey(roomId), String(value || ""));
      return true;
    } catch (err) {
      return false;
    }
  }

  function clearRoomDraft(roomId) {
    try {
      window.localStorage.removeItem(roomDraftStorageKey(roomId));
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadStoredActiveRoomId() {
    try {
      return Number(window.localStorage.getItem(activeRoomStorageKey()) || 0);
    } catch (err) {
      return 0;
    }
  }

  function saveStoredActiveRoomId(roomId) {
    try {
      if (!roomId) {
        window.localStorage.removeItem(activeRoomStorageKey());
        return true;
      }
      window.localStorage.setItem(activeRoomStorageKey(), String(Number(roomId) || 0));
      return true;
    } catch (err) {
      return false;
    }
  }

  function loadStoredActiveNavId() {
    try {
      return String(window.localStorage.getItem(activeNavStorageKey()) || "");
    } catch (err) {
      return "";
    }
  }

  function saveStoredActiveNavId(navId) {
    try {
      if (!navId) {
        window.localStorage.removeItem(activeNavStorageKey());
        return true;
      }
      window.localStorage.setItem(activeNavStorageKey(), String(navId));
      return true;
    } catch (err) {
      return false;
    }
  }

  function syncFocusModeUI() {
    var button = q("btn-focus-mode");
    if (!button) {
      return;
    }
    button.textContent = state.focusMode ? "Foco on" : "Foco off";
    button.classList.toggle("active", !!state.focusMode);
  }

  function isChatContextForcedOpen() {
    var room = activeRoom();
    return !!isAppsLabRoom(room);
  }

  function syncChatContextUI() {
    var button = q("btn-context-toggle");
    var title = q("conversation-context-title");
    var collapsed = !!state.chatContextCollapsed && !isChatContextForcedOpen() && !isSusView();
    if (title) {
      title.textContent = isAppsLabRoom(activeRoom()) ? "Central tecnica do app" : "Resumo da area";
    }
    if (!button) {
      return;
    }
    if (isSusView()) {
      button.disabled = true;
      button.textContent = "Indisponivel";
      button.classList.remove("active");
      return;
    }
    if (isChatContextForcedOpen()) {
      button.disabled = true;
      button.textContent = "Fixo";
      button.classList.remove("active");
      return;
    }
    button.disabled = false;
    button.textContent = collapsed ? "Expandir" : "Recolher";
    button.classList.toggle("active", !collapsed);
  }

  function syncUtilityStripUI() {
    var button = q("btn-utility-toggle");
    var collapsed = !!state.utilityStripCollapsed && !!state.compactLayout;
    document.body.classList.toggle("utility-strip-collapsed", collapsed);
    if (!button) {
      return;
    }
    button.classList.toggle("hidden", !state.compactLayout);
    button.classList.toggle("active", !!state.compactLayout && !collapsed);
    button.setAttribute("aria-expanded", collapsed ? "false" : "true");
    button.textContent = collapsed ? "Atalhos" : "Esconder";
  }

  function syncComposerDraftHint() {
    var hint = q("composer-draft-hint");
    var input = q("composer-input");
    var room = activeRoom();
    var raw = input ? String(input.value || "") : "";
    if (!hint) {
      return;
    }
    hint.classList.remove("has-draft");
    hint.classList.remove("is-empty");
    if (state.editingMessage) {
      hint.textContent = "Editando mensagem selecionada";
      return;
    }
    if (!room || isHubView()) {
      hint.textContent = "Escolhe uma sala pra rascunhar";
      hint.classList.add("is-empty");
      return;
    }
    if (raw.trim()) {
      hint.textContent = "Rascunho salvo local // " + raw.length + " chars";
      hint.classList.add("has-draft");
      return;
    }
    hint.textContent = "Sem rascunho salvo";
    hint.classList.add("is-empty");
  }

  function restoreComposerDraft() {
    var input = q("composer-input");
    var room = activeRoom();
    if (!input || state.editingMessage) {
      syncComposerDraftHint();
      return;
    }
    input.value = room ? loadRoomDraft(room.id) : "";
    syncComposerDraftHint();
  }

  function persistComposerDraft() {
    var input = q("composer-input");
    var room = activeRoom();
    if (!input || !room || state.editingMessage) {
      syncComposerDraftHint();
      return;
    }
    if (String(input.value || "").trim()) {
      saveRoomDraft(room.id, input.value);
    } else {
      clearRoomDraft(room.id);
    }
    syncComposerDraftHint();
  }

  function appInstallLabel() {
    return state.appInstalled ? "app detectado" : "app nao detectado";
  }

  function appInstallCopy() {
    if (!state.appInstalled) {
      return "Se tu ja instalou, usa Abrir no app. Quando o protocolo responder, o painel confirma isso aqui.";
    }
    if (!state.appInstallSeenAt) {
      return "O protocolo do app ja respondeu nesta maquina.";
    }
    return "Detectado nesta maquina em " + formatDateTime(state.appInstallSeenAt) + ".";
  }

  function syncAppInstallStatus() {
    var label = appInstallLabel();
    if (q("login-app-status-pill")) {
      q("login-app-status-pill").textContent = label;
    }
    if (q("apps-app-status-pill")) {
      q("apps-app-status-pill").textContent = label;
    }
    if (q("app-center-install-pill")) {
      q("app-center-install-pill").textContent = label;
    }
    if (q("app-center-install-label")) {
      q("app-center-install-label").textContent = state.appInstalled ? "App detectado nesta maquina" : "App ainda nao detectado";
    }
    if (q("app-center-install-copy")) {
      q("app-center-install-copy").textContent = appInstallCopy();
    }
    if (q("app-center-version-pill")) {
      q("app-center-version-pill").textContent = UNIVERSALD_META.shortVersion;
    }
    if (q("app-center-build-copy")) {
      q("app-center-build-copy").textContent = UNIVERSALD_META.version;
    }
    if (q("app-center-sha-copy")) {
      q("app-center-sha-copy").textContent = "SHA " + UNIVERSALD_META.sha256.slice(0, 16) + "...";
    }
    if (q("app-center-size-copy")) {
      q("app-center-size-copy").textContent = UNIVERSALD_META.sizeLabel;
    }
    if (q("app-center-origin-copy")) {
      q("app-center-origin-copy").textContent = "Distribuido por " + window.location.origin;
    }
  }

  function renderContextEmptyCard(title, copy, pillsMarkup, actionsMarkup) {
    return "<article class='context-empty-card'>" +
      "<div class='context-empty-copy'>" +
        "<strong>" + esc(title || "Sem conteudo ainda") + "</strong>" +
        "<p>" + esc(copy || "Essa area ainda esta quieta.") + "</p>" +
      "</div>" +
      (pillsMarkup ? "<div class='context-empty-pills'>" + pillsMarkup + "</div>" : "") +
      (actionsMarkup ? "<div class='context-empty-actions'>" + actionsMarkup + "</div>" : "") +
    "</article>";
  }

  function renderDirectHubDeck() {
    var wrap = q("direct-hub-deck");
    var rooms;
    var unreadTotal;
    var onlinePeers;
    var hottest;
    var summaryActions;
    if (!wrap) {
      return;
    }
    if (state.activeNavId !== "diretas") {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    rooms = directRoomsSorted();
    wrap.classList.remove("hidden");
    if (!rooms.length) {
      wrap.innerHTML = renderContextEmptyCard(
        "Diretas vazias",
        "Ainda nao existe DM aberta. Vai em membros, escolhe alguem e puxa a primeira conversa privada da base.",
        "<span class='ghost-pill'>pv</span><span class='ghost-pill'>1x1</span><span class='ghost-pill'>sem plateia</span>",
        "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='members'>Ver membros</button>" +
        "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='search'>Abrir busca</button>"
      );
      return;
    }
    unreadTotal = rooms.reduce(function(sum, room) {
      return sum + Number(state.unread[String(room.id)] || 0);
    }, 0);
    onlinePeers = rooms.filter(function(room) {
      var peer = directPeerProfile(room);
      return !!(peer && peer.online);
    }).length;
    hottest = rooms[0];
    summaryActions =
      "<span class='ghost-pill'>" + rooms.length + " conversas</span>" +
      "<span class='ghost-pill'>" + onlinePeers + " online</span>" +
      "<span class='ghost-pill'>" + unreadTotal + " novas</span>" +
      "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='members'>Membros</button>" +
      "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='search'>Busca</button>";
    wrap.innerHTML =
      "<div class='context-deck-head'>" +
        "<div class='context-deck-copy'>" +
          "<span class='eyebrow'>diretas</span>" +
          "<strong>Conversas privadas vivas</strong>" +
          "<p>Retoma rapido quem te chamou, quem esta online e onde a conversa esquentou.</p>" +
        "</div>" +
        "<div class='context-deck-actions'>" + summaryActions + "</div>" +
      "</div>" +
      "<div class='direct-hub-summary'>" +
        "<article class='insight-card'><span>diretas</span><strong>" + rooms.length + "</strong><p>conversas privadas abertas</p></article>" +
        "<article class='insight-card'><span>na fila</span><strong>" + unreadTotal + "</strong><p>mensagens novas te esperando</p></article>" +
        "<article class='insight-card'><span>online</span><strong>" + onlinePeers + "</strong><p>contatos vivos no mapa agora</p></article>" +
        "<article class='insight-card'><span>mais quente</span><strong>" + esc(hottest ? displayRoomName(hottest) : "sem foco") + "</strong><p>" + esc(hottest && hottest.lastMessageAt ? formatRelative(hottest.lastMessageAt) : "sem troca recente") + "</p></article>" +
      "</div>" +
      "<div class='direct-hub-grid'>" +
        rooms.slice(0, 4).map(function(room) {
          var peer = directPeerProfile(room);
          var unread = Number(state.unread[String(room.id)] || 0);
          var latest = state.latestByRoom[String(room.id)] || {};
          var statusCopy = personStatusCopy(peer);
          var liveCopy = peer && peer.online ? (peer.status || "online") : "offline";
          var preview = latest.body || (latest.attachment ? "[anexo] " + latest.attachment.name : "Sem mensagem recente.");
          var latestType = latest.attachment ? (latest.attachment.kind || "anexo") : "texto";
          return "<article class='direct-card'>" +
            "<div class='direct-card-head'>" +
              directPeerAvatarMarkup(room) +
              "<div class='direct-peer-copy'>" +
                "<strong>" + esc(displayRoomName(room)) + "</strong>" +
                "<span>" + esc(liveCopy + (statusCopy ? (" // " + statusCopy) : "")) + "</span>" +
              "</div>" +
              (unread ? "<span class='ghost-pill'>" + unread + " novas</span>" : "<span class='ghost-pill'>" + esc(room.lastMessageAt ? formatRelative(room.lastMessageAt) : "sem rastro") + "</span>") +
            "</div>" +
            "<div class='direct-card-copy'>" +
              "<strong>" + esc(peer && peer.online ? "na base agora" : "ultimas trocas") + "</strong>" +
              "<p>" + esc(preview.slice(0, 140)) + "</p>" +
            "</div>" +
            "<div class='inline-row direct-card-meta'>" +
              "<span class='ghost-pill'>" + esc(latestType) + "</span>" +
              "<span class='ghost-pill'>" + esc(room.lastMessageAt ? formatRelative(room.lastMessageAt) : "sem rastro") + "</span>" +
              (latest.attachment && latest.attachment.sizeBytes ? "<span class='ghost-pill'>" + esc(bytesLabel(latest.attachment.sizeBytes)) + "</span>" : "") +
            "</div>" +
            "<div class='inline-row'>" +
              "<button class='btn btn-ghost' type='button' data-action='jump-room' data-room-id='" + Number(room.id) + "'>" + (unread ? "Voltar pras novas" : "Abrir DM") + "</button>" +
              "<button class='btn btn-ghost' type='button' data-action='focus-composer'>Responder</button>" +
              "<button class='btn btn-ghost' type='button' data-action='open-user-profile' data-user-id='" + Number(room.peerUserId || 0) + "'>Perfil</button>" +
            "</div>" +
          "</article>";
        }).join("") +
      "</div>";
  }

  function renderRoomSpotlight() {
    var wrap = q("room-spotlight");
    var room = activeRoom();
    var attachments;
    var sectionLabel;
    if (!wrap) {
      return;
    }
    if (!room || isAppsLabRoom(room) || isHubView() || room.category !== "media") {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    attachments = filteredRoomAttachments().slice(0, 4);
    wrap.classList.remove("hidden");
    sectionLabel = room && room.slug === "fotos" ? "galeria" : "biblioteca";
    if (!attachments.length) {
      wrap.innerHTML = renderContextEmptyCard(
        "Vitrine vazia",
        "Essa area ainda nao tem midia em destaque. Solta foto, video ou arquivo e a vitrine acorda.",
        "<span class='ghost-pill'>" + esc(sectionLabel) + "</span><span class='ghost-pill'>midia</span>",
        "<button class='btn btn-ghost' type='button' data-action='trigger-attach'>Anexar agora</button>" +
        "<button class='btn btn-ghost' type='button' data-action='focus-composer'>Escrever junto</button>"
      );
      return;
    }
    wrap.innerHTML =
      "<div class='context-deck-head'>" +
        "<div class='context-deck-copy'>" +
          "<span class='eyebrow'>" + esc(sectionLabel) + "</span>" +
          "<strong>Midia em destaque da sala</strong>" +
          "<p>As ultimas pecas que mais representam essa area agora.</p>" +
        "</div>" +
        "<div class='context-deck-actions'>" +
          "<span class='ghost-pill'>" + attachments.length + " na vitrine</span>" +
          "<span class='ghost-pill'>" + esc(bytesLabel(totalAttachmentBytes(attachments))) + "</span>" +
          "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='files'>Biblioteca</button>" +
        "</div>" +
      "</div>" +
      "<div class='spotlight-grid'>" +
        attachments.map(function(message) {
          var attachment = message.attachment;
          var attachmentUrl = safeAttachmentUrl(attachment && attachment.url);
          var mediaMarkup = "<div class='media-file-fallback'><strong>" + esc((attachment.extension || attachment.kind || "arquivo").toUpperCase()) + "</strong><span>" + esc(bytesLabel(attachment.sizeBytes)) + "</span></div>";
          if (attachmentUrl && attachment.kind === "image") {
            mediaMarkup = "<button type='button' class='spotlight-media' data-action='preview-attachment' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'><img alt='" + esc(attachment.name) + "' src='" + esc(attachmentUrl) + "' /></button>";
          } else if (attachmentUrl && attachment.kind === "video") {
            mediaMarkup = "<button type='button' class='spotlight-media' data-action='preview-attachment' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'><video preload='metadata' muted src='" + esc(attachmentUrl) + "'></video></button>";
          } else if (attachment.kind === "audio") {
            mediaMarkup = "<div class='spotlight-media'><div class='media-file-fallback'><strong>AUDIO</strong><span>" + esc(bytesLabel(attachment.sizeBytes)) + "</span></div></div>";
          }
          return "<article class='spotlight-card'>" +
            "<div class='spotlight-card-head'>" +
              "<div class='spotlight-card-copy'>" +
                "<strong>" + esc(attachment.name) + "</strong>" +
                "<p>" + esc((message.body || displayRoomDescription(room)).slice(0, 110)) + "</p>" +
              "</div>" +
              "<span class='ghost-pill'>" + esc(message.authorName || "alguem") + "</span>" +
            "</div>" +
            mediaMarkup +
            "<div class='inline-row'>" +
              "<span class='ghost-pill'>" + esc((attachment.kind || "arquivo") + " // " + formatRelative(message.createdAt)) + "</span>" +
              "<span class='ghost-pill'>" + esc(bytesLabel(attachment.sizeBytes)) + "</span>" +
              "<button class='btn btn-ghost' type='button' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Origem</button>" +
            "</div>" +
          "</article>";
        }).join("") +
      "</div>";
  }

  function renderMediaInsights(room, attachments, total) {
    var wrap = q("media-insights");
    var totals;
    var totalBytes = 0;
    if (!wrap) {
      return;
    }
    if (!room || isHubView() || isAppsLabRoom(room)) {
      wrap.innerHTML = "";
      wrap.classList.add("hidden");
      return;
    }
    totals = {
      image: 0,
      video: 0,
      audio: 0,
      file: 0
    };
    currentRoomAttachments().forEach(function(message) {
      var kind = String(message.attachment && message.attachment.kind || "file");
      if (!totals[kind] && totals[kind] !== 0) {
        totals[kind] = 0;
      }
      totals[kind] += 1;
      totalBytes += Number(message.attachment && message.attachment.sizeBytes || 0);
    });
    wrap.classList.remove("hidden");
    wrap.innerHTML =
      "<article class='insight-card'><span>visiveis</span><strong>" + attachments.length + "</strong><p>itens batendo no filtro atual</p></article>" +
      "<article class='insight-card'><span>biblioteca</span><strong>" + total + "</strong><p>midias salvas nessa area</p></article>" +
      "<article class='insight-card'><span>fotos</span><strong>" + totals.image + "</strong><p>prints e imagens da sala</p></article>" +
      "<article class='insight-card'><span>videos</span><strong>" + totals.video + "</strong><p>clips e videozinhos enviados</p></article>" +
      "<article class='insight-card'><span>arquivos</span><strong>" + (totals.file + totals.audio) + "</strong><p>docs, audios e pacotes</p></article>" +
      "<article class='insight-card'><span>peso</span><strong>" + esc(bytesLabel(totalBytes)) + "</strong><p>biblioteca total carregada</p></article>";
  }

  function renderMembersInsights() {
    var wrap = q("members-insights");
    var items = filteredPresenceItems();
    var onlineCount;
    var adminCount;
    var vipCount;
    var withStatusCount;
    if (!wrap) {
      return;
    }
    onlineCount = items.filter(function(item) { return item.online; }).length;
    adminCount = items.filter(function(item) { return item.role === "admin" || item.role === "owner"; }).length;
    vipCount = items.filter(function(item) { return item.role === "vip"; }).length;
    withStatusCount = items.filter(function(item) { return !!personStatusCopy(item); }).length;
    wrap.innerHTML =
      "<article class='insight-card'><span>visiveis</span><strong>" + items.length + "</strong><p>membros no recorte atual</p></article>" +
      "<article class='insight-card'><span>online</span><strong>" + onlineCount + "</strong><p>ativos agora na base</p></article>" +
      "<article class='insight-card'><span>admin</span><strong>" + adminCount + "</strong><p>owner + admins no mapa</p></article>" +
      "<article class='insight-card'><span>vip</span><strong>" + vipCount + "</strong><p>vip carregados no filtro</p></article>" +
      "<article class='insight-card'><span>status</span><strong>" + withStatusCount + "</strong><p>com status customizado visivel</p></article>";
  }

  function renderDashboardPulse() {
    var wrap = q("dashboard-pulse");
    var room = activeRoom();
    var roomMessages = currentRoomMessages();
    var viewerStats = currentViewerLocalStats();
    var latest = room ? (state.latestByRoom[String(room.id)] || null) : null;
    var latestLabel = latest
      ? ((latest.authorName || "base") + " // " + formatRelative(latest.createdAt))
      : "sem rastro recente";
    if (!wrap) {
      return;
    }
    wrap.innerHTML =
      "<article class='insight-card'><span>nao lidas</span><strong>" + unreadTotalCount() + "</strong><p>mensagens fora da tua sala ativa</p></article>" +
      "<article class='insight-card'><span>pedidos</span><strong>" + Number(state.pendingJoinCount || 0) + "</strong><p>gente pedindo guarida na base</p></article>" +
      "<article class='insight-card'><span>viewer</span><strong>" + viewerStats.messageCount + "</strong><p>mensagens tuas carregadas no mapa</p></article>" +
      "<article class='insight-card'><span>midia</span><strong>" + viewerStats.attachmentCount + "</strong><p>anexos teus no recorte local</p></article>" +
      "<article class='insight-card'><span>sala ativa</span><strong>" + (roomMessages.length || 0) + "</strong><p>" + esc(room ? displayRoomName(room) : "sem sala em foco") + " // " + esc(latestLabel) + "</p></article>" +
      "<article class='insight-card'><span>conexao</span><strong>" + esc(state.connectionStatus || "syncing") + "</strong><p>" + esc(window.Notification ? notificationPermissionLabel() : "push indisponivel") + "</p></article>";
  }

  function renderAppCenter() {
    if (!q("app-center-changelog-list")) {
      return;
    }
    if (q("login-app-size-pill")) {
      q("login-app-size-pill").textContent = UNIVERSALD_META.sizeLabel;
    }
    q("app-center-version-pill").textContent = UNIVERSALD_META.shortVersion;
    q("app-center-build-copy").textContent = UNIVERSALD_META.version;
    q("app-center-sha-copy").textContent = "SHA " + UNIVERSALD_META.sha256.slice(0, 16) + "...";
    q("app-center-size-copy").textContent = UNIVERSALD_META.sizeLabel;
    q("app-center-protocol-pill").textContent = "universald://";
    q("app-center-changelog-list").innerHTML = UNIVERSALD_META.changelog.map(function(item) {
      return "<li>" + esc(item) + "</li>";
    }).join("");
    syncAppInstallStatus();
  }

  function setConnectionStatus(status) {
    state.connectionStatus = String(status || "syncing");
    renderConnectionStatus();
  }

  function renderConnectionStatus() {
    var badge = q("connection-badge");
    var label = "sync";
    var tone = "badge-sync";
    if (!badge) {
      return;
    }
    if (state.connectionStatus === "live") {
      label = "live";
      tone = "badge badge-sync live";
    } else if (state.connectionStatus === "unstable") {
      label = "oscilando";
      tone = "badge badge-sync unstable";
    } else if (state.connectionStatus === "offline") {
      label = "offline";
      tone = "badge badge-sync offline";
    } else {
      label = "sync";
      tone = "badge badge-sync";
    }
    badge.className = tone;
    badge.textContent = label;
  }

  function syncThemePresetState() {
    var activeTheme = String(q("profile-theme") && q("profile-theme").value || state.viewer && state.viewer.theme || "matrix");
    Array.prototype.slice.call(document.querySelectorAll("[data-theme-preset]")).forEach(function(button) {
      button.classList.toggle("active", button.getAttribute("data-theme-preset") === activeTheme);
    });
  }

  function syncBannerPresetState() {
    var activeBanner = String(q("profile-banner-preset") && q("profile-banner-preset").value || state.viewer && state.viewer.bannerPreset || "grid");
    Array.prototype.slice.call(document.querySelectorAll("[data-banner-preset]")).forEach(function(button) {
      button.classList.toggle("active", button.getAttribute("data-banner-preset") === activeBanner);
    });
  }

  function copyText(text, successText, failText) {
    if (!navigator.clipboard || !navigator.clipboard.writeText) {
      toast(failText || "Nao consegui copiar isso agora.", "warn");
      return Promise.resolve(false);
    }
    return navigator.clipboard.writeText(String(text || "")).then(function() {
      toast(successText || "Copiado.", "ok");
      return true;
    }).catch(function() {
      toast(failText || "Nao consegui copiar isso agora.", "warn");
      return false;
    });
  }

  function generatePanelPassword(length) {
    var size = Math.max(10, Math.min(32, Number(length || 18)));
    var alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%&*?";
    var output = "";
    var index;
    for (index = 0; index < size; index++) {
      output += alphabet.charAt(Math.floor(Math.random() * alphabet.length));
    }
    return output;
  }

  function roomModeLabel(room, mode) {
    if (isSusView()) {
      return "sus";
    }
    if (isMembersHub()) {
      return "membros";
    }
    if (isDirectHubEmpty()) {
      return "diretas";
    }
    if (!room) {
      return "idle";
    }
    if (mode === "dm") {
      return "direta";
    }
    if (mode === "locked") {
      return "senha";
    }
    if (mode === "vip") {
      return "vip";
    }
    if (mode === "admin") {
      return "admin";
    }
    return "aberta";
  }

  function roomPulseLabel(room) {
    if (isSusView()) {
      return "anomalia";
    }
    var total = currentRoomMessages().length;
    if (isMembersHub()) {
      return "radar do grupo";
    }
    if (isDirectHubEmpty()) {
      return "diretas prontas";
    }
    if (!room) {
      return "sem foco";
    }
    if (total >= 28) {
      return "pilhada";
    }
    if (total >= 10) {
      return "aquecida";
    }
    if (total >= 1) {
      return "de boa";
    }
    return "silenciosa";
  }

  function notificationPermissionLabel() {
    if (!window.Notification) {
      return "Push indisponivel";
    }
    if (Notification.permission === "granted") {
      return "Push ativo";
    }
    if (Notification.permission === "denied") {
      return "Push negado";
    }
    return "Ativar push";
  }

  function syncBrowserNotifyButton() {
    var button = q("btn-browser-notify");
    var stateLabel = "unsupported";
    if (!button) {
      return;
    }
    if (window.Notification && Notification.permission === "granted") {
      stateLabel = "granted";
    } else if (window.Notification && Notification.permission === "denied") {
      stateLabel = "denied";
    } else if (window.Notification) {
      stateLabel = "default";
    }
    button.textContent = notificationPermissionLabel();
    button.disabled = !window.Notification;
    button.classList.toggle("active", !!window.Notification && Notification.permission === "granted");
    button.setAttribute("data-push-state", stateLabel);
    button.title = !window.Notification
      ? "Teu navegador nao suporta push por aqui."
      : (Notification.permission === "granted"
        ? "Push ligado nesse dispositivo."
        : (Notification.permission === "denied"
          ? "Push bloqueado nas permissoes do navegador."
          : "Clica para pedir permissao de push."));
  }

  function syncMobileActionButtons() {
    var profileUtility = q("btn-profile-utility");
    var logoutUtility = q("btn-logout-utility");
    var inspectorToggle = q("btn-inspector-toggle");
    if (!profileUtility || !logoutUtility) {
      return;
    }
    profileUtility.classList.toggle("hidden", !state.compactLayout);
    logoutUtility.classList.toggle("hidden", !state.compactLayout);
    if (inspectorToggle) {
      inspectorToggle.classList.toggle("hidden", !!state.isMobile);
    }
  }

  function directPeerAvatarMarkup(room) {
    var peer = directPeerProfile(room);
    var name = displayRoomName(room);
    var avatarUrl = safeAvatarUrl(peer && peer.avatarUrl);
    if (avatarUrl) {
      return "<img class='direct-peer-avatar' alt='" + esc(name) + "' src='" + esc(avatarUrl) + "' />";
    }
    return "<div class='direct-peer-avatar'>" + esc(initials(name)) + "</div>";
  }

  function negoDashboardTip(room, mode) {
    if (isSusView()) {
      return "Bah... se tu insiste nessa aba, depois aguenta o que vier do outro lado.";
    }
    if (hasUploadingPendingUploads()) {
      return "Bah, segura um pouco que ainda tem upload subindo no motor.";
    }
    if (hasErroredPendingUploads()) {
      return "Capaz, tem upload falhando ali. Arruma isso antes de largar a mensagem.";
    }
    if (isMembersHub()) {
      return "Abre um perfil, puxa uma DM e faz o painel parecer vivo.";
    }
    if (isDirectHubEmpty()) {
      return "Te manda nos membros e abre a primeira direta sem cerimônia.";
    }
    if (!room) {
      return "Escolhe uma sala na esquerda e eu te acompanho no resto.";
    }
    if (mode === "locked") {
      return "Essa porta tem senha. Entra direito ou fica no corredor, tchê.";
    }
    if (mode === "vip") {
      return "Sala fina, acesso fino. Sem convite ela não abre nem com reza.";
    }
    if (room.slug === "apps-lab") {
      return "Apps Lab é pra testar ideia sem poluir o resto. Usa a nota local e gera senha forte.";
    }
    if (room.category === "media") {
      return "Arrasta arquivo pro chat que eu deixo a biblioteca dessa sala mais redonda.";
    }
    if (room.category === "ia") {
      return "Larga tua pergunta sem medo. Se tiver drama técnico eu mastigo junto.";
    }
    if (room.scope === "dm") {
      return "DM boa é direta, clara e sem mandar textão torto.";
    }
    if (!currentRoomMessages().length) {
      return "Ninguém falou aqui ainda. Manda a primeira e acorda esse canto.";
    }
    return "Sala " + roomPulseLabel(room) + ". Mantém o papo afiado e sem flood.";
  }

  function nudgeConversation() {
    pulseClass(q("conversation-panel"), "panel-shift");
  }

  function currentRoomPins() {
    return state.pinnedByRoom[String(state.activeRoomId)] || [];
  }

  function directPeerProfile(room) {
    if (!room || room.scope !== "dm" || !room.peerUserId) {
      return null;
    }
    return presenceByUserId(room.peerUserId);
  }

  function latestMessagesMap(items) {
    var map = {};
    (items || []).forEach(function(message) {
      if (message && Number(message.roomId) > 0) {
        map[String(message.roomId)] = message;
      }
    });
    return map;
  }

  function normalizeMentionToken(value) {
    return normalizeText(value).replace(/\s+/g, "");
  }

  function mentionTargets() {
    var tokens = {};
    var role = normalizeText(state.viewer && state.viewer.role);
    [state.viewer && state.viewer.username, state.viewer && state.viewer.displayName].forEach(function(name) {
      var clean = normalizeMentionToken(name);
      if (!clean) {
        return;
      }
      tokens[clean] = true;
      tokens[clean.replace(/[\s_.-]+/g, "")] = true;
      tokens[clean.replace(/\s+/g, "-")] = true;
      tokens[clean.replace(/\s+/g, ".")] = true;
      tokens[clean.replace(/\s+/g, "_")] = true;
    });
    tokens.todos = true;
    tokens.all = true;
    if (role === "owner") {
      tokens.owner = true;
      tokens.admin = true;
      tokens.vip = true;
    }
    if (role === "admin") {
      tokens.admin = true;
      tokens.vip = true;
    }
    if (role === "vip") {
      tokens.vip = true;
    }
    return tokens;
  }

  function mentionHit(token) {
    var clean = normalizeMentionToken(String(token || "").replace(/^@+/, ""));
    var targets = mentionTargets();
    return !!targets[clean];
  }

  function messageMentionsViewer(message) {
    var text = String(message && message.body || "");
    var matched = text.match(/@[A-Za-zÀ-ÿ0-9_.-]+/g) || [];
    return matched.some(mentionHit);
  }

  function renderMessageBodyMarkup(message) {
    var text = String(message && message.body || "");
    var urlPattern = /((https?:\/\/|www\.)[^\s<]+)/gi;
    if (!text) {
      return "";
    }
    var escaped = esc(text).replace(/\n/g, "<br>");
    escaped = escaped.replace(/(^|[\s>])(@[A-Za-zÀ-ÿ0-9_.-]+)/g, function(match, prefix, token) {
      var className = "mention-pill" + (mentionHit(token) ? " mine" : "");
      return prefix + "<span class='" + className + "'>" + token + "</span>";
    });
    escaped = escaped.replace(urlPattern, function(match) {
      var href = /^https?:\/\//i.test(match) ? match : ("https://" + match);
      return "<a class='message-link' href='" + esc(href) + "' target='_blank' rel='noreferrer'>" + esc(match) + "</a>";
    });
    return "<p class='message-body'>" + escaped + "</p>";
  }

  function canManageMessage(message) {
    if (!state.viewer || !message) {
      return false;
    }
    return canManageOwnedContent(message.authorId);
  }

  function canPinMessage(message) {
    return canManageMessage(message);
  }

  function canManageOwnedContent(ownerId) {
    if (!state.viewer || !ownerId) {
      return false;
    }
    return Number(ownerId) === Number(state.viewer.id) || state.viewer.role === "owner" || state.viewer.role === "admin";
  }

  function syncLocalSocialFlags(userId, key, enabled) {
    var numericUserId = Number(userId);
    var listName = key === "blocked" ? "blockedUserIds" : "mutedUserIds";
    var flagName = key === "blocked" ? "blockedByViewer" : "mutedByViewer";
    state[listName] = (state[listName] || []).filter(function(item) {
      return Number(item) !== numericUserId;
    });
    if (enabled) {
      state[listName].push(numericUserId);
    }
    state.online = (state.online || []).map(function(item) {
      if (Number(item.userId) === numericUserId) {
        item[flagName] = enabled;
      }
      return item;
    });
    if (key === "blocked") {
      Object.keys(state.messagesByRoom).forEach(function(roomId) {
        state.messagesByRoom[roomId] = (state.messagesByRoom[roomId] || []).map(function(message) {
          if (Number(message.authorId) === numericUserId) {
            message.blockedByViewer = enabled;
          }
          return message;
        });
      });
      Object.keys(state.latestByRoom).forEach(function(roomId) {
        if (state.latestByRoom[roomId] && Number(state.latestByRoom[roomId].authorId) === numericUserId) {
          state.latestByRoom[roomId].blockedByViewer = enabled;
        }
      });
    }
    if (state.selectedMember && Number(state.selectedMember.user.userId) === numericUserId) {
      state.selectedMember.user[flagName] = enabled;
      state.selectedMember.canDm = !state.selectedMember.user.blockedByViewer && !state.selectedMember.user.hasBlockedViewer && Number(state.selectedMember.user.userId) !== Number(state.viewer && state.viewer.id) && state.selectedMember.user.role !== "ai";
    }
  }

  function detectMobile() {
    var agent = String((window.navigator && window.navigator.userAgent) || "").toLowerCase();
    var viewport = window.visualViewport || null;
    var layoutWidth = viewport && viewport.width ? viewport.width : window.innerWidth;
    var wasMobile = !!state.isMobile;
    var wasCompact = !!state.compactLayout;
    state.isMobile = layoutWidth <= 820 || /android|iphone|ipad|mobile|opera mini|windows phone/.test(agent);
    state.compactLayout = state.isMobile || layoutWidth <= 1080;
    document.body.classList.toggle("mobile", state.isMobile);
    document.body.classList.toggle("compact", state.compactLayout);
    document.body.classList.toggle("mobile-clean", state.isMobile);
    if (state.compactLayout) {
      if (!wasCompact) {
        document.body.classList.remove("sidebar-open");
        document.body.classList.remove("inspector-open");
        state.utilityStripCollapsed = loadUtilityStripCollapsed();
      }
      if (state.isMobile && !wasMobile) {
        state.utilityStripCollapsed = true;
        state.chatContextCollapsed = true;
      }
      if (state.isMobile && document.body.classList.contains("composer-focus")) {
        state.utilityStripCollapsed = true;
        state.chatContextCollapsed = true;
      }
      document.body.classList.toggle("sidebar-collapsed", !document.body.classList.contains("sidebar-open"));
      document.body.classList.toggle("inspector-collapsed", !document.body.classList.contains("inspector-open"));
    } else {
      document.body.classList.remove("sidebar-open");
      document.body.classList.toggle("sidebar-collapsed", !!state.desktopSidebarCollapsed);
      document.body.classList.remove("inspector-open");
      document.body.classList.toggle("inspector-collapsed", !!state.desktopInspectorCollapsed);
      if (!state.desktopInspectorCollapsed) {
        document.body.classList.add("inspector-open");
      }
      state.utilityStripCollapsed = false;
    }
    q("device-pill").textContent = state.isMobile ? "mobile mode" : (state.compactLayout ? "compact mode" : "desktop mode");
    applyFocusMode();
    applyChatContextState();
    syncUtilityStripUI();
    syncMobileActionButtons();
    syncBackdrop();
    syncPeekButtons();
  }

  function applyFocusMode() {
    document.body.classList.toggle("focus-mode", !!state.focusMode);
    if (!state.focusMode) {
      syncFocusModeUI();
      return;
    }
    document.body.classList.remove("sidebar-open");
    document.body.classList.remove("inspector-open");
    document.body.classList.add("sidebar-collapsed");
    document.body.classList.add("inspector-collapsed");
    syncFocusModeUI();
    syncBackdrop();
    syncPeekButtons();
  }

  function applyChatContextState() {
    var collapsed = !!state.chatContextCollapsed && !isChatContextForcedOpen() && !isSusView();
    document.body.classList.toggle("chat-context-collapsed", collapsed);
    syncChatContextUI();
  }

  function toggleFocusMode() {
    state.focusMode = !state.focusMode;
    saveFocusMode();
    applyFocusMode();
    if (!state.focusMode) {
      detectMobile();
    }
    renderHeaderAndRoomState();
    toast(state.focusMode ? "Modo foco ligado. Agora o chat respira." : "Modo foco desligado. A base voltou completa.", "ok");
  }

  function toggleChatContext() {
    if (isChatContextForcedOpen() || isSusView()) {
      applyChatContextState();
      return;
    }
    state.chatContextCollapsed = !state.chatContextCollapsed;
    saveChatContextCollapsed();
    applyChatContextState();
    toast(state.chatContextCollapsed ? "Contexto recolhido. O chat ganhou mais area." : "Contexto expandido. Os cards voltaram pro topo.", "ok");
  }

  function toggleUtilityStrip() {
    if (!state.compactLayout) {
      state.utilityStripCollapsed = false;
      syncUtilityStripUI();
      return;
    }
    state.utilityStripCollapsed = !state.utilityStripCollapsed;
    saveUtilityStripCollapsed();
    syncUtilityStripUI();
  }

  function collapseMobileChrome(forceContext) {
    if (!state.isMobile) {
      return;
    }
    if (!state.utilityStripCollapsed) {
      state.utilityStripCollapsed = true;
      saveUtilityStripCollapsed();
      syncUtilityStripUI();
    }
    if (forceContext && !isChatContextForcedOpen() && !isSusView() && !state.chatContextCollapsed) {
      state.chatContextCollapsed = true;
      saveChatContextCollapsed();
      applyChatContextState();
    }
  }

  function setComposerFocusMode(active) {
    if (!state.compactLayout) {
      document.body.classList.remove("composer-focus");
      return;
    }
    if (active) {
      if (state.isMobile) {
        state.utilityStripCollapsed = true;
        state.chatContextCollapsed = true;
        syncUtilityStripUI();
        applyChatContextState();
      }
      closeSidebar();
      closeInspector();
      document.body.classList.add("composer-focus");
      if (state.isMobile) {
        try {
          q("panel-view").scrollTop = 0;
          window.scrollTo(0, 0);
        } catch (e) {}
      }
      return;
    }
    document.body.classList.remove("composer-focus");
    if (state.isMobile) {
      syncUtilityStripUI();
      applyChatContextState();
    }
  }

  function syncTransientLayoutState() {
    var loginVisible = !q("login-view").classList.contains("hidden");
    var panelVisible = !q("panel-view").classList.contains("hidden");
    var hasModal = !!document.querySelector(".modal:not(.hidden)");
    var drawerOpen;
    if (!panelVisible) {
      document.body.classList.remove("sidebar-open");
      document.body.classList.remove("inspector-open");
    }
    drawerOpen = !!(panelVisible && state.compactLayout && (document.body.classList.contains("sidebar-open") || document.body.classList.contains("inspector-open")));
    document.body.classList.toggle("modal-open", !!(hasModal && (loginVisible || panelVisible)));
    document.body.classList.toggle("drawer-open", drawerOpen);
  }

  function blurInteractiveTarget(target) {
    if (!target || typeof target.blur !== "function") {
      return;
    }
    try {
      target.blur();
    } catch (e) {}
  }

  function syncBackdrop() {
    var visible = state.compactLayout && (document.body.classList.contains("sidebar-open") || document.body.classList.contains("inspector-open"));
    q("mobile-backdrop").classList.toggle("hidden", !visible);
    syncTransientLayoutState();
  }

  function syncPeekButtons() {
    if (state.compactLayout) {
      q("btn-sidebar-peek").classList.add("hidden");
      q("btn-inspector-peek").classList.add("hidden");
      return;
    }
    q("btn-sidebar-peek").classList.toggle("hidden", !document.body.classList.contains("sidebar-collapsed"));
    q("btn-inspector-peek").classList.toggle("hidden", !document.body.classList.contains("inspector-collapsed"));
  }

  function ensureActiveNavVisible(force) {
    var list = document.querySelector("#channel-panel .sidebar-scroll");
    var active = document.querySelector(".room-item.active");
    var activeKey;
    var listRect;
    var itemRect;
    var nextTop;
    if (!active) {
      return;
    }
    activeKey = String(active.getAttribute("data-nav-id") || active.getAttribute("data-room-id") || "");
    if (!force && state.lastEnsuredNavKey === activeKey) {
      return;
    }
    if (!list) {
      state.lastEnsuredNavKey = activeKey;
      return;
    }
    listRect = list.getBoundingClientRect();
    itemRect = active.getBoundingClientRect();
    if (!force && itemRect.top >= listRect.top + 8 && itemRect.bottom <= listRect.bottom - 8) {
      state.lastEnsuredNavKey = activeKey;
      return;
    }
    nextTop = list.scrollTop + (itemRect.top - listRect.top) - Math.max(10, Math.round((listRect.height - itemRect.height) / 2));
    list.scrollTop = Math.max(0, nextTop);
    state.lastEnsuredNavKey = activeKey;
  }

  function openSidebar() {
    if (!state.compactLayout) {
      state.desktopSidebarCollapsed = false;
      document.body.classList.remove("sidebar-collapsed");
      syncPeekButtons();
      return;
    }
    if (state.isMobile) {
      state.utilityStripCollapsed = true;
      syncUtilityStripUI();
    }
    closeInspector();
    document.body.classList.add("sidebar-open");
    document.body.classList.remove("sidebar-collapsed");
    syncBackdrop();
    ensureActiveNavVisible(true);
  }

  function closeSidebar() {
    if (!state.compactLayout) {
      state.desktopSidebarCollapsed = true;
      document.body.classList.add("sidebar-collapsed");
      syncPeekButtons();
      return;
    }
    document.body.classList.remove("sidebar-open");
    document.body.classList.add("sidebar-collapsed");
    syncBackdrop();
  }

  function openInspector() {
    if (state.compactLayout) {
      if (state.isMobile) {
        state.utilityStripCollapsed = true;
        syncUtilityStripUI();
      }
      closeSidebar();
      document.body.classList.remove("inspector-collapsed");
      document.body.classList.add("inspector-open");
      syncBackdrop();
      return;
    }
    state.desktopInspectorCollapsed = false;
    document.body.classList.remove("inspector-collapsed");
    document.body.classList.add("inspector-open");
    syncPeekButtons();
    syncTransientLayoutState();
  }

  function closeInspector() {
    if (!state.compactLayout) {
      state.desktopInspectorCollapsed = true;
      document.body.classList.add("inspector-collapsed");
      document.body.classList.remove("inspector-open");
      syncPeekButtons();
      return;
    }
    document.body.classList.remove("inspector-open");
    document.body.classList.add("inspector-collapsed");
    syncBackdrop();
  }

  function submitComposer() {
    var form = q("composer-form");
    if (!form) {
      return;
    }
    if (typeof form.requestSubmit === "function") {
      form.requestSubmit();
      return;
    }
    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
  }

  function pulseClass(element, className) {
    if (!element) {
      return;
    }
    element.classList.remove(className);
    void element.offsetWidth;
    element.classList.add(className);
  }

  function wait(ms) {
    return new Promise(function(resolve) {
      window.setTimeout(resolve, ms);
    });
  }

  function playTone(kind) {
    var AudioCtor = window.AudioContext || window.webkitAudioContext;
    var sets = {
      ok: [620, 760],
      warn: [420, 320],
      err: [180, 140],
      sus: [97, 131, 83]
    };
    if (!state.audioEnabled || !AudioCtor || !sets[kind]) {
      return;
    }
    try {
      if (!state.audioContext) {
        state.audioContext = new AudioCtor();
      }
      if (state.audioContext.state === "suspended") {
        state.audioContext.resume();
      }
      sets[kind].forEach(function(frequency, index) {
        var oscillator = state.audioContext.createOscillator();
        var gain = state.audioContext.createGain();
        var startAt = state.audioContext.currentTime + (index * 0.06);
        oscillator.type = kind === "err" ? "sawtooth" : (kind === "sus" ? "square" : "triangle");
        oscillator.frequency.setValueAtTime(frequency, startAt);
        gain.gain.setValueAtTime(0.0001, startAt);
        gain.gain.exponentialRampToValueAtTime(kind === "warn" ? 0.018 : (kind === "sus" ? 0.016 : 0.028), startAt + 0.015);
        gain.gain.exponentialRampToValueAtTime(0.0001, startAt + (kind === "sus" ? 0.22 : 0.14));
        oscillator.connect(gain);
        gain.connect(state.audioContext.destination);
        oscillator.start(startAt);
        oscillator.stop(startAt + (kind === "sus" ? 0.24 : 0.16));
      });
    } catch (e) {}
  }

  function maybeVibrate(pattern) {
    if (!state.isMobile || !navigator || typeof navigator.vibrate !== "function") {
      return;
    }
    try {
      navigator.vibrate(pattern);
    } catch (e) {}
  }

  function pushActivity(text, tone, meta) {
    state.activity.unshift({
      id: Date.now() + "-" + Math.round(Math.random() * 100000),
      text: text,
      tone: tone || "ok",
      at: new Date().toISOString(),
      roomId: meta && meta.roomId ? Number(meta.roomId) : 0,
      messageId: meta && meta.messageId ? Number(meta.messageId) : 0
    });
    if (state.activity.length > 40) {
      state.activity = state.activity.slice(0, 40);
    }
    renderActivity();
    renderNotificationCenter();
  }

  function toast(text, tone, meta) {
    var stack = q("toast-stack");
    var icon = "\uD83D\uDCE1";
    var label = "Painel Dief";
    var toastKey = String(tone || "ok") + "::" + String(text || "");
    var existing;
    if (!stack) {
      return;
    }
    if (state.lastToastKey === toastKey && Date.now() - Number(state.lastToastAt || 0) < 1400) {
      return;
    }
    state.lastToastKey = toastKey;
    state.lastToastAt = Date.now();
    if (tone === "ok") {
      label = "Bah, ficou tinindo";
      icon = "\uD83D\uDE0E";
    } else if (tone === "warn") {
      label = "Opa, segura essa";
      icon = "\u26A0\uFE0F";
    } else if (tone === "err") {
      label = "Deu ruim no circuito";
      icon = "\uD83D\uDE35";
    }
    var item = document.createElement("div");
    item.className = "toast " + (tone || "ok");
    item.innerHTML =
      "<span class='toast-icon'>" + esc(icon) + "</span>" +
      "<div class='toast-copy'><strong>" + esc(label) + "</strong><span>" + esc(text) + "</span></div>" +
      "<button type='button' class='toast-close' aria-label='Fechar aviso'>x</button>";
    existing = stack.querySelectorAll(".toast");
    while (existing.length >= 4) {
      if (existing[0] && existing[0].parentNode) {
        existing[0].parentNode.removeChild(existing[0]);
      }
      existing = stack.querySelectorAll(".toast");
    }
    stack.appendChild(item);
    item.addEventListener("click", function(event) {
      var close = event.target && event.target.closest ? event.target.closest(".toast-close") : null;
      if (!close) {
        openNotificationCenter();
        return;
      }
      if (item.parentNode) {
        item.parentNode.removeChild(item);
      }
    });
    pushActivity(label + ". " + text, tone || "ok", meta);
    playTone(tone || "ok");
    if (tone === "warn") {
      maybeVibrate([22, 58, 22]);
    } else if (tone === "err") {
      maybeVibrate([42, 85, 42]);
    }
    window.setTimeout(function() {
      if (item.parentNode) {
        item.parentNode.removeChild(item);
      }
    }, 3800);
  }

  function showSusOmen(text) {
    var omen = q("sus-omen");
    if (!omen) {
      return;
    }
    omen.textContent = text || "Tem alguma coisa olhando de volta.";
    omen.classList.remove("hidden");
    omen.classList.remove("active");
    window.setTimeout(function() {
      omen.classList.add("active");
    }, 20);
    if (state.susOmenTimer) {
      window.clearTimeout(state.susOmenTimer);
    }
    state.susOmenTimer = window.setTimeout(function() {
      omen.classList.remove("active");
      window.setTimeout(function() {
        omen.classList.add("hidden");
      }, 180);
    }, 1750);
  }

  function setButtonBusy(target, busy, busyText, defaultText) {
    var button = typeof target === "string" ? q(target) : target;
    if (!button) {
      return;
    }
    if (defaultText) {
      button.dataset.defaultText = defaultText;
    } else if (!button.dataset.defaultText) {
      button.dataset.defaultText = button.textContent;
    }
    if (busy) {
      button.disabled = true;
      button.classList.add("is-busy");
      button.setAttribute("aria-busy", "true");
      button.textContent = busyText || "Processando...";
      return;
    }
    button.classList.remove("is-busy");
    button.removeAttribute("aria-busy");
    button.disabled = false;
    button.textContent = button.dataset.defaultText || defaultText || button.textContent;
  }

  function toggleAudio() {
    state.audioEnabled = !state.audioEnabled;
    q("btn-audio-toggle").textContent = state.audioEnabled ? "Som on" : "Som off";
    toast(state.audioEnabled ? "Som do painel ligado." : "Som do painel silenciado.", state.audioEnabled ? "ok" : "warn");
  }

  function notifyBrowser(title, body) {
    if (!window.Notification || Notification.permission !== "granted") {
      return;
    }
    try {
      new Notification(title, { body: body });
    } catch (e) {}
  }

  function totalUnreadCount() {
    return Object.keys(state.unread || {}).reduce(function(sum, key) {
      return sum + Number(state.unread[key] || 0);
    }, 0);
  }

  function updateDocumentTitle() {
    var unread = totalUnreadCount();
    var room = activeRoom();
    var base = "Painel Dief";
    if (room) {
      base += " // " + displayRoomName(room);
    }
    document.title = unread > 0 ? ("(" + unread + ") " + base) : base;
  }

  function requestNotificationsPermission() {
    syncBrowserNotifyButton();
  }

  function activityFeedItems(limit) {
    var latestItems = Object.keys(state.latestByRoom || {}).map(function(roomId) {
      var message = state.latestByRoom[roomId];
      var room = roomById(roomId);
      if (!message || !room) {
        return null;
      }
      return {
        id: "room-" + roomId + "-" + Number(message.id || 0),
        at: message.createdAt,
        tone: room.scope === "dm" ? "ok" : "warn",
        roomId: Number(room.id || 0),
        messageId: Number(message.id || 0),
        text: displayRoomName(room) + " // " + (message.authorName || "alguem") + " // " + ((message.body || (message.attachment ? message.attachment.name : "sem texto")).slice(0, 96))
      };
    }).filter(Boolean);
    var upcomingEvent = (state.events || []).slice(0, 2).map(function(event) {
      return {
        id: "event-" + Number(event.id || 0),
        at: event.startsAt,
        tone: "ok",
        roomId: Number(event.roomId || 0),
        messageId: 0,
        text: "Agenda // " + event.title + " // " + formatRelative(event.startsAt)
      };
    });
    return latestItems
      .concat(upcomingEvent, state.activity || [])
      .filter(Boolean)
      .sort(function(a, b) {
        return new Date(b.at).getTime() - new Date(a.at).getTime();
      })
      .slice(0, Math.max(1, Number(limit || 8)));
  }

  function openNotificationCenter() {
    if (state.isMobile) {
      closeSidebar();
      closeInspector();
      state.utilityStripCollapsed = true;
      syncUtilityStripUI();
    }
    renderNotificationCenter();
    q("notify-modal").classList.remove("hidden");
    q("notify-modal").scrollTop = 0;
    state.notificationCenterOpen = true;
    syncTransientLayoutState();
  }

  function clearNotificationCenter() {
    state.activity = [];
    renderActivity();
    renderNotificationCenter();
    playTone("ok");
  }

  function syncNotificationCenterButton() {
    var button = q("btn-notify-center");
    var count = activityFeedItems(12).length;
    if (!button) {
      return;
    }
    button.textContent = count ? ("Avisos " + count) : "Avisos";
    button.classList.toggle("active", count > 0);
  }

  function renderNotificationCenter() {
    var list = q("notification-list");
    var items = activityFeedItems(24);
    if (!list) {
      return;
    }
    if (!items.length) {
      list.innerHTML = "<div class='empty-state'>Sem alerta recente. Quando a base se mexer, tudo aparece aqui.</div>";
      syncNotificationCenterButton();
      return;
    }
    list.innerHTML = items.map(function(item) {
      var jumpButton = item.roomId
        ? "<button class='btn btn-ghost' type='button' data-action='notification-jump' data-room-id='" + Number(item.roomId) + "'" + (item.messageId ? " data-message-id='" + Number(item.messageId) + "'" : "") + ">Abrir</button>"
        : "";
      return "<article class='notification-card " + esc(item.tone || "ok") + "'>" +
        "<div class='notification-card-head'>" +
          "<strong>" + esc(formatDateTime(item.at)) + "</strong>" +
          "<span class='ghost-pill'>" + esc(item.tone || "ok") + "</span>" +
        "</div>" +
        "<p>" + esc(item.text || "Sem detalhe.") + "</p>" +
        (jumpButton ? "<div class='inline-row notification-card-actions'>" + jumpButton + "</div>" : "") +
      "</article>";
    }).join("");
    syncNotificationCenterButton();
  }

  function handleBrowserNotifyToggle() {
    var permissionRequest;
    if (!window.Notification) {
      toast("Teu navegador nao libera push por aqui.", "warn");
      syncBrowserNotifyButton();
      return;
    }
    if (Notification.permission === "granted") {
      toast("Push do navegador ja esta ligado nesse device.", "ok");
      syncBrowserNotifyButton();
      return;
    }
    if (Notification.permission === "denied") {
      toast("O navegador travou as notificacoes. Libera nas permissoes do site.", "warn");
      syncBrowserNotifyButton();
      return;
    }
    try {
      permissionRequest = Notification.requestPermission(function(permission) {
        syncBrowserNotifyButton();
        if (permission === "granted") {
          toast("Push do navegador ligado. Agora o painel te cutuca.", "ok");
          return;
        }
        toast("Push nao foi liberado. O painel segue no visual e no som.", "warn");
      });
    } catch (requestErr) {
      syncBrowserNotifyButton();
      toast("Nao consegui pedir permissao de push agora.", "err");
      return;
    }
    if (!permissionRequest || typeof permissionRequest.then !== "function") {
      return;
    }
    permissionRequest.then(function(permission) {
      syncBrowserNotifyButton();
      if (permission === "granted") {
        toast("Push do navegador ligado. Agora o painel te cutuca.", "ok");
        return;
      }
      toast("Push nao foi liberado. O painel segue no visual e no som.", "warn");
    }).catch(function() {
      syncBrowserNotifyButton();
      toast("Nao consegui pedir permissao de push agora.", "err");
    });
  }

  async function apiFetch(path, options) {
    var opts = options || {};
    var headers = opts.headers || {};
    var storedSessionId = readStoredSessionId();
    var isFormData = typeof FormData !== "undefined" && opts.body instanceof FormData;
    if (!isFormData && !headers["Content-Type"]) {
      headers["Content-Type"] = "application/json";
    }
    if (storedSessionId && !headers["X-Panel-Session"]) {
      headers["X-Panel-Session"] = storedSessionId;
    }
    var response = await fetch(path, {
      method: opts.method || "GET",
      credentials: "same-origin",
      headers: headers,
      body: opts.body
    });
    var data = {};
    try {
      data = await response.json();
    } catch (e) {}
    if (!response.ok) {
      var requestId = response.headers.get("X-Request-ID") || (data && data.requestId) || "";
      if (response.status === 401 && isPanelSessionManagedPath(path)) {
        clearStoredSessionId();
        resetSession();
      }
      var err = new Error((data && data.error) || "falha inesperada");
      if (requestId) {
        err.requestId = requestId;
      }
      throw err;
    }
    return data;
  }

  function setViewMode(mode) {
    var body = document.body;
    if (!body) {
      return;
    }
    body.classList.toggle("login-mode", mode === "login");
    body.classList.toggle("panel-mode", mode === "panel");
  }

  function showLogin(message) {
    state.emergencyMode = false;
    document.body.classList.remove("emergency-mode");
    clearPanelEmergencyLayout();
    setViewMode("login");
    q("login-view").classList.remove("hidden");
    q("panel-view").classList.add("hidden");
    document.body.classList.remove("modal-open", "drawer-open", "utility-strip-collapsed", "sidebar-open", "inspector-open", "focus-mode");
    document.title = "Painel Dief";
    q("login-view").querySelector(".login-stage").classList.remove("login-error-pulse", "login-success-pulse");
    q("login-view").scrollTop = 0;
    state.focusMode = false;
    state.utilityStripCollapsed = false;
    syncTransientLayoutState();
    syncBackdrop();
    syncPeekButtons();
    window.scrollTo(0, 0);
    if (message) {
      q("login-error").classList.remove("hidden");
      q("login-error").textContent = message;
    } else {
      q("login-error").classList.add("hidden");
      q("login-error").textContent = "";
    }
    window.setTimeout(function() {
      var loginInput = q("login-input");
      if (!loginInput) {
        return;
      }
      try {
        loginInput.focus({ preventScroll: true });
      } catch (e) {
        loginInput.focus();
      }
    }, 40);
  }

  function showPanel() {
    state.emergencyMode = false;
    document.body.classList.remove("emergency-mode");
    clearPanelEmergencyLayout();
    setViewMode("panel");
    q("login-view").classList.add("hidden");
    q("panel-view").classList.remove("hidden");
    q("panel-view").scrollTop = 0;
    document.body.classList.remove("modal-open", "drawer-open", "sidebar-open", "inspector-open");
    detectMobile();
    if (state.isMobile) {
      state.utilityStripCollapsed = true;
      state.chatContextCollapsed = true;
    }
    closeSidebar();
    closeInspector();
    applyChatContextState();
    syncUtilityStripUI();
    syncTransientLayoutState();
    syncBackdrop();
    syncPeekButtons();
    window.scrollTo(0, 0);
    ensurePanelSurfaceVisible();
    maybeOpenGuide();
    renderAppCenter();
  }

  function ensurePanelSurfaceVisible() {
    window.setTimeout(function() {
      var panelView = q("panel-view");
      var workspace = q("workspace");
      var topbar = q("topbar");
      var conversation = q("conversation-panel");
      var workspaceHeight;
      var conversationHeight;
      if (!panelView || panelView.classList.contains("hidden") || !workspace || !conversation) {
        return;
      }
      workspaceHeight = Math.round(workspace.getBoundingClientRect().height || 0);
      conversationHeight = Math.round(conversation.getBoundingClientRect().height || 0);
      if (workspaceHeight >= 180 && conversationHeight >= 180) {
        clearPanelEmergencyLayout();
        return;
      }
      state.focusMode = false;
      state.utilityStripCollapsed = false;
      state.chatContextCollapsed = false;
      document.body.classList.remove(
        "focus-mode",
        "modal-open",
        "drawer-open",
        "sidebar-open",
        "inspector-open",
        "utility-strip-collapsed",
        "chat-context-collapsed"
      );
      detectMobile();
      closeSidebar();
      closeInspector();
      applyFocusMode();
      applyChatContextState();
      syncUtilityStripUI();
      syncTransientLayoutState();
      syncBackdrop();
      syncPeekButtons();
      workspaceHeight = Math.round(workspace.getBoundingClientRect().height || 0);
      conversationHeight = Math.round(conversation.getBoundingClientRect().height || 0);
      if (workspaceHeight >= 180 && conversationHeight >= 180) {
        clearPanelEmergencyLayout();
        return;
      }
      if (state.isMobile || state.compactLayout) {
        return;
      }
      activatePanelEmergencyLayout();
      safeScrollIntoView(topbar, { block: "start", behavior: "auto" });
    }, 90);
  }

  function resetSession() {
    clearStoredSessionId();
    state.viewer = null;
    state.rooms = [];
    state.roomAccess = {};
    state.online = [];
    state.typing = [];
    state.recentLogs = [];
    state.latestByRoom = {};
    state.events = [];
    state.polls = [];
    state.roomVersions = {};
    state.blockedUserIds = [];
    state.mutedUserIds = [];
    state.users = [];
    state.joinRequests = [];
    state.joinRequestsLoaded = false;
    state.pendingJoinCount = 0;
    state.activeRoomId = 0;
    state.activeNavId = "chat-geral";
    state.messagesByRoom = {};
    state.pinnedByRoom = {};
    state.unread = {};
    state.latestUnreadByRoom = {};
    state.pendingAttachment = null;
    state.pendingUploads = [];
    state.replyTarget = null;
    state.editingMessage = null;
    state.selectedMember = null;
    state.version = 0;
    state.filter = "all";
    state.inspectorTab = "overview";
    state.searchQuery = "";
    state.searchResult = null;
    state.mediaFilter = "all";
    state.mediaSearch = "";
    state.mediaPreview = null;
    state.highlightMessageId = 0;
    state.typingSent = false;
    state.roomSearch = "";
    state.compactLayout = false;
    state.focusMode = false;
    state.utilityStripCollapsed = false;
    state.connectionStatus = "offline";
    state.lastStreamAt = 0;
    state.favoriteNavIds = [];
    state.desktopSidebarCollapsed = false;
    state.desktopInspectorCollapsed = false;
    updateDocumentTitle();
    if (state.stream) {
      state.stream.close();
      state.stream = null;
    }
    if (state.heartbeatTimer) {
      window.clearInterval(state.heartbeatTimer);
      state.heartbeatTimer = null;
    }
    if (state.typingIdleTimer) {
      window.clearTimeout(state.typingIdleTimer);
      state.typingIdleTimer = null;
    }
    document.body.classList.remove("sidebar-open");
    document.body.classList.remove("inspector-open");
    document.body.classList.remove("sidebar-collapsed");
    document.body.classList.remove("inspector-collapsed");
    document.body.classList.remove("compact");
    document.body.classList.remove("focus-mode");
    q("room-filter-input").value = "";
    syncBackdrop();
    syncPeekButtons();
    renderConnectionStatus();
    syncAppInstallStatus();
    showLogin("Tua sessao caiu. Faz o login de novo.");
  }

  async function tryBootstrap() {
    try {
      var bootstrap = await apiFetch("/api/panel/bootstrap");
      if (bootstrap && bootstrap.sessionId) {
        storeSessionId(bootstrap.sessionId);
      }
      applyBootstrap(bootstrap);
      showPanel();
    } catch (err) {
      showLogin("");
    }
  }

  function applyTheme() {
    if (!state.viewer) {
      return;
    }
    document.body.setAttribute("data-theme", state.viewer.theme || "matrix");
    document.documentElement.style.setProperty("--accent", state.viewer.accentColor || "#7bff00");
  }

  function pickDefaultRoomId() {
    var i;
    for (i = 0; i < state.rooms.length; i++) {
      var access = accessForRoom(state.rooms[i]);
      if (access === "open" || access === "admin") {
        return state.rooms[i].id;
      }
    }
    return state.rooms.length ? state.rooms[0].id : 0;
  }

  function applyBootstrap(bootstrap) {
    var preferredRoomId;
    var preferredNavId;
    var preferredRoom;
    var preferredNav;
    state.viewer = bootstrap.viewer;
    state.rooms = bootstrap.rooms || [];
    state.roomAccess = bootstrap.roomAccess || {};
    state.online = bootstrap.online || [];
    state.typing = bootstrap.typing || [];
    state.recentLogs = bootstrap.recentLogs || [];
    state.latestByRoom = latestMessagesMap(bootstrap.latestMessages || []);
    state.events = bootstrap.events || [];
    state.polls = [];
    state.roomVersions = bootstrap.roomVersions || {};
    state.blockedUserIds = (bootstrap.blockedUserIds || []).map(Number);
    state.mutedUserIds = (bootstrap.mutedUserIds || []).map(Number);
    state.version = Number(bootstrap.version || 0);
    state.previousOnlineMap = onlineMap(state.online);
    state.appsNotesDraft = loadAppsNotes();
    state.favoriteNavIds = loadFavoriteNavIds();
    state.focusMode = loadFocusMode();
    state.chatContextCollapsed = loadChatContextCollapsed();
    state.utilityStripCollapsed = loadUtilityStripCollapsed();
    if (!state.viewer || state.viewer.role !== "owner") {
      state.users = [];
      state.joinRequests = [];
      state.joinRequestsLoaded = false;
      state.pendingJoinCount = 0;
    }
    preferredRoomId = loadStoredActiveRoomId();
    preferredNavId = loadStoredActiveNavId();
    syncPendingAttachmentAlias();
    q("room-filter-input").value = state.roomSearch;
    preferredRoom = preferredRoomId ? roomById(preferredRoomId) : null;
    if (!roomById(state.activeRoomId) && preferredRoom) {
      state.activeRoomId = preferredRoom.id;
    }
    if (!roomById(state.activeRoomId)) {
      state.activeRoomId = pickDefaultRoomId();
    }
    if (hasPrimaryNavId(preferredNavId)) {
      preferredNav = navDefinition(preferredNavId);
      if (preferredNav.id === "sus") {
        state.activeNavId = "sus";
        state.activeRoomId = 0;
      } else if (preferredNav.id === "diretas" && !primaryDirectRoom()) {
        state.activeNavId = "diretas";
        state.activeRoomId = 0;
      } else if (navRoom(preferredNav)) {
        state.activeNavId = preferredNav.id;
        state.activeRoomId = navRoom(preferredNav).id;
      } else {
        state.activeNavId = navIdForRoom(roomById(state.activeRoomId));
      }
    } else {
      state.activeNavId = navIdForRoom(roomById(state.activeRoomId));
    }
    applyTheme();
    requestNotificationsPermission();
    renderShell();
    updateDocumentTitle();
    applyFocusMode();
    applyChatContextState();
    syncUtilityStripUI();
    renderAppCenter();
    ensurePanelSurfaceVisible();
    setConnectionStatus("syncing");
    if (state.activeRoomId) {
      selectRoom(state.activeRoomId, true);
    } else if (state.activeNavId === "sus") {
      selectPrimaryNav("sus", true);
    }
    if (state.viewer && state.viewer.role === "owner") {
      loadUsers();
      loadJoinRequests();
    }
    setupStream();
    setupHeartbeat();
  }

  function renderShell() {
    renderViewer();
    renderRails();
    renderRoomLegend();
    renderRooms();
    renderPresenceMini();
    renderPresenceSidebar();
    renderHeaderAndRoomState();
    renderDashboard();
    renderDashboardPulse();
    renderDirectHubDeck();
    renderRoomSpotlight();
    renderAppsLabDeck();
    renderSusPanel();
    renderPinnedStrip();
    renderUnreadBanner();
    renderPendingAttachment();
    renderReplyChip();
    renderEditChip();
    renderActivity();
    renderNotificationCenter();
    renderEvents();
    renderMoments();
    renderFiles();
    renderFavorites();
    renderPolls();
    renderLogs();
    renderJoinRequests();
    renderAdminPanels();
    renderInspectorTabs();
    renderTypingIndicator();
    renderMediaPreview();
    syncComposerDraftHint();
    syncDocumentTitle();
    syncUtilityStripUI();
    syncPeekButtons();
  }

  function renderViewer() {
    if (!state.viewer) {
      return;
    }
    var viewerStatusText = personStatusCopy(state.viewer);
    q("viewer-name").textContent = state.viewer.displayName || state.viewer.username;
    q("viewer-role").textContent = state.viewer.role || "member";
    q("viewer-bio").textContent = state.viewer.bio || "Sem bio definida.";
    q("viewer-theme-pill").textContent = themeLabel(state.viewer.theme || "matrix");
    q("viewer-status-pill").textContent = state.viewer.status || "online";
    q("viewer-status-copy").textContent = viewerStatusText || "Sem status customizado.";
    q("viewer-status-copy").classList.toggle("hidden", !viewerStatusText);
    q("btn-audio-toggle").textContent = state.audioEnabled ? "Som on" : "Som off";
    syncFocusModeUI();
    syncAppInstallStatus();
    syncBrowserNotifyButton();
    renderConnectionStatus();
    q("overview-viewer-name").textContent = state.viewer.displayName || state.viewer.username;
    q("overview-viewer-copy").textContent = viewerStatusText || ((state.viewer.role || "member") + " // " + themeLabel(state.viewer.theme || "matrix"));
    q("composer-upload-hint").textContent = "/help /logs /matrix /online // upload ate " + uploadLimitLabel();

    var avatar = q("viewer-avatar");
    var avatarUrl = safeAvatarUrl(state.viewer.avatarUrl);
    if (avatarUrl) {
      avatar.src = avatarUrl;
      avatar.alt = "Avatar de " + (state.viewer.displayName || state.viewer.username);
    } else {
      avatar.removeAttribute("src");
      avatar.alt = initials(state.viewer.displayName || state.viewer.username);
    }
  }

  function renderRails() {
    var nav = navDefinition(state.activeNavId);
    q("active-filter-label").textContent = (nav && nav.label || "chat geral").toLowerCase();
  }

  function renderRoomLegend() {
    q("room-legend").innerHTML =
      "<span class='room-kind kind-chat'>chat</span>" +
      "<span class='room-kind kind-dm'>dm</span>" +
      "<span class='room-kind kind-media'>midia</span>" +
      "<span class='room-kind kind-dev'>ia</span>" +
      "<span class='room-kind kind-admin'>admin</span>" +
      "<span class='room-kind kind-sus'>sus</span>";
  }

  function syncDocumentTitle() {
    var room = activeRoom();
    var nav = navDefinition(state.activeNavId);
    var unread = 0;
    Object.keys(state.unread).forEach(function(roomId) {
      unread += Number(state.unread[roomId] || 0);
    });
    document.title = (unread > 0 ? "(" + unread + ") " : "") + "Painel Dief" + " | " + (room ? displayRoomName(room) : (nav && nav.label || "Painel"));
  }

  function renderRooms() {
    var list = q("room-list");
    var totalUnread = 0;
    var visible = sortNavItems(PRIMARY_NAV.filter(function(nav) {
      return navVisible(nav) && navSearchAllows(nav);
    }));

    list.innerHTML = "";
    q("room-count").textContent = String(visible.length);

    state.rooms.forEach(function(room) {
      totalUnread += Number(state.unread[String(room.id)] || 0);
    });
    q("stat-unread").textContent = String(totalUnread);

    if (!visible.length) {
      list.innerHTML = "<div class='empty-state'>Nenhuma das " + PRIMARY_NAV.length + " areas bateu com esse filtro.</div>";
      return;
    }

    visible.forEach(function(nav) {
      var room = navRoom(nav);
      var access = accessForRoom(room);
      var unread = room ? Number(state.unread[String(room.id)] || 0) : 0;
      var badge = "";
      var button = document.createElement("button");
      var kind = room ? roomKind(room) : (nav.kind === "dm" ? "dm" : "chat");
      var category = room ? (room.category || "") : (nav.kind || "chat");
      var scope = room ? (room.scope || "public") : "hub";
      var title = room ? displayRoomName(room) : nav.label;
      var description = room ? (room.lastMessagePreview || displayRoomDescription(room)) : nav.copy;
      var subline = room
        ? (roomKindLabel(room) + " // " + description)
        : ("hub // " + nav.copy);
      var accent = room ? roomPulseLabel(room) : (nav.accent || "base");
      var theme = nav.theme || roomKindLabel(room);
      if (access === "locked") {
        badge = "<span class='badge badge-lock'>senha</span>";
      } else if (access === "vip") {
        badge = "<span class='badge badge-vip'>vip</span>";
      } else if (access === "admin") {
        badge = "<span class='badge badge-admin'>admin</span>";
      }
      button.type = "button";
      button.className = "room-item" + (state.activeNavId === nav.id ? " active" : "");
      if (nav.kind === "sus") {
        button.className += " sus-room";
      }
      if (isFavoriteNavId(nav.id)) {
        button.className += " favorite";
      }
      button.setAttribute("data-nav-id", nav.id);
      button.setAttribute("data-category", category);
      button.setAttribute("data-scope", scope);
      button.innerHTML =
        "<div class='room-item-top'>" +
          "<div class='room-item-title'>" +
            "<span class='room-icon'>" + esc(room ? displayRoomIcon(room) : nav.icon) + "</span>" +
            "<div class='room-title-copy'>" +
              "<strong>" + esc(title) + "</strong>" +
              "<span>" + esc(subline) + "</span>" +
              "<div class='room-copy-badges'><span class='room-theme-pill'>" + esc(theme) + "</span><span class='room-theme-pill soft'>" + esc(accent) + "</span></div>" +
            "</div>" +
          "</div>" +
          "<div class='room-meta'>" + badge + (unread > 0 ? "<span class='unread-badge'>" + unread + "</span>" : "") + "</div>" +
        "</div>";
      list.appendChild(button);
    });
    renderDirectHubDeck();
  }

  function renderPresenceMini() {
    var list = q("presence-mini");
    var onlineItems = state.online.filter(function(item) { return item.online; });
    var items = onlineItems.slice(0, 6);
    q("online-count").textContent = String(onlineItems.length);
    q("presence-count").textContent = onlineItems.length + " on";
    list.innerHTML = "";
    if (!items.length) {
      list.innerHTML = "<div class='empty-state'>Ninguem colou no painel ainda.</div>";
      return;
    }
    items.forEach(function(person) {
      var card = document.createElement("div");
      card.className = "presence-chip";
      card.innerHTML =
        "<strong>" + esc(person.displayName || person.username) + "</strong>" +
        "<span>" + esc((person.role || "member") + " // " + (person.status || "online")) + "</span>";
      list.appendChild(card);
    });
  }

  function renderPresenceSidebar() {
    var list = q("presence-sidebar");
    var summary = q("members-summary");
    var items = filteredPresenceItems();
    var onlineVisible = items.filter(function(item) { return item.online; }).length;
    q("members-status-pill").textContent = onlineVisible + " online";
    renderMembersInsights();
    list.innerHTML = "";
    if (!items.length) {
      if (summary) {
        summary.textContent = "Nenhum membro bate com teu filtro agora.";
      }
      list.innerHTML = "<div class='empty-state'>Sem membros carregados nesse recorte. Limpa o filtro ou muda o cargo.</div>";
      return;
    }
    if (summary) {
      summary.textContent = items.length + " membros visiveis // " + onlineVisible + " online // filtro " + (state.memberRoleFilter === "all" ? "geral" : state.memberRoleFilter);
    }
    items.forEach(function(person) {
      var canDM = state.viewer && Number(person.userId) !== Number(state.viewer.id) && person.role !== "ai";
      var socialBadges = [];
      var avatarUrl = safeAvatarUrl(person.avatarUrl);
      var statusCopy = personStatusCopy(person);
      var liveRoom = person.online && person.roomId ? roomById(person.roomId) : null;
      if (person.blockedByViewer) {
        socialBadges.push("bloqueado");
      }
      if (person.mutedByViewer) {
        socialBadges.push("silenciado");
      }
      if (person.hasBlockedViewer) {
        socialBadges.push("te bloqueou");
      }
      var card = document.createElement("div");
      card.className = "presence-item";
      card.innerHTML =
        "<button class='identity-trigger presence-identity' type='button' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>" +
          (avatarUrl ? "<img class='presence-avatar' alt='" + esc(person.displayName || person.username) + "' src='" + esc(avatarUrl) + "' />" : "<div class='presence-avatar'>" + esc(initials(person.displayName || person.username)) + "</div>") +
        "</button>" +
        "<div class='presence-text'>" +
          "<div class='presence-headline'>" +
            "<button type='button' class='identity-trigger identity-name' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>" + esc(person.displayName || person.username) + "</button>" +
            "<span class='presence-room-copy'>" + esc(liveRoom ? displayRoomName(liveRoom) : (person.online ? "na base" : "offline")) + "</span>" +
          "</div>" +
          "<div class='inline-row presence-pills'>" +
            "<span class='ghost-pill'>" + esc(person.role || "member") + "</span>" +
            "<span class='ghost-pill'>" + esc(person.online ? (person.status || "online") : "offline") + "</span>" +
            (statusCopy ? "<span class='ghost-pill ghost-pill-soft'>" + esc(statusCopy) + "</span>" : "") +
          "</div>" +
          (socialBadges.length ? "<span>" + esc(socialBadges.join(" // ")) + "</span>" : "") +
        "</div>" +
        "<div class='presence-actions'>" +
          "<button class='btn btn-ghost' type='button' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>Perfil</button>" +
          (canDM && !person.blockedByViewer && !person.hasBlockedViewer ? "<button class='btn btn-ghost' type='button' data-action='open-dm' data-user-id='" + Number(person.userId) + "'>DM</button>" : "") +
        "</div>" +
        "<span class='presence-dot " + (person.online ? "online" : "offline") + "'></span>";
      list.appendChild(card);
    });
  }

  function renderHeaderAndRoomState() {
    var room = activeRoom();
    var access = accessForRoom(room);
    var peer = directPeerProfile(room);
    var stateCard = q("room-state");
    var btnAction = q("btn-room-action");
    var btnRoomFavorite = q("btn-room-favorite");
    var composerDisabled = false;
    var memberText = "";
    var hubMode = isHubView();

    renderDashboardPulse();

    btnRoomFavorite.disabled = !state.activeNavId;
    btnRoomFavorite.textContent = isFavoriteNavId(state.activeNavId) ? "Area fixa" : "Fixar area";
    btnRoomFavorite.classList.toggle("active", isFavoriteNavId(state.activeNavId));

    q("composer-form").classList.toggle("hidden", hubMode);
    q("conversation-panel").classList.toggle("apps-lab-mode", !!isAppsLabRoom(room));
    q("conversation-panel").classList.toggle("sus-mode", isSusView());
    q("message-stream").classList.remove("hidden");
    q("typing-indicator").classList.remove("hidden");
    q("sus-panel").classList.add("hidden");
    q("pins-strip").classList.remove("hidden");
    q("unread-banner").classList.remove("hidden");
    applyChatContextState();

    if (isSusView()) {
      q("room-icon").textContent = "??";
      q("room-title").textContent = "???";
      q("room-description").textContent = "Aba pegadinha, estranha e mal encarada. Se clicar demais, tu some dali pro outro lado.";
      q("overview-room-name").textContent = "???";
      q("overview-room-copy").textContent = "Area suspeita sem chat normal. Aqui so tem misterio, aviso e um botao maldito.";
      q("room-members-pill").textContent = "sus";
      q("stat-version").textContent = "v" + String(state.version || 0);
      q("room-type-badge").className = "badge kind-sus";
      q("room-type-badge").textContent = "sus";
      q("room-access-badge").className = "badge badge-lock";
      q("room-access-badge").textContent = "na mexe";
      q("composer-form").classList.add("hidden");
      q("message-stream").classList.add("hidden");
      q("typing-indicator").classList.add("hidden");
      stateCard.classList.add("hidden");
      q("sus-panel").classList.remove("hidden");
      applyChatContextState();
      return;
    }

    if (isMembersHub()) {
      q("room-icon").textContent = "ON";
      q("room-title").textContent = "Membros";
      q("room-description").textContent = "Hub do grupo com presenca online, atalhos de DM e leitura limpa.";
      q("overview-room-name").textContent = "Membros";
      q("overview-room-copy").textContent = "Lista viva de quem esta online e pronto pra conversar.";
      q("room-members-pill").textContent = state.online.filter(function(item) { return item.online; }).length + " on";
      q("stat-version").textContent = "v" + String(state.version || 0);
      q("room-type-badge").className = "badge kind-admin";
      q("room-type-badge").textContent = "hub";
      q("room-access-badge").className = "badge badge-live";
      q("room-access-badge").textContent = "grupo";
      stateCard.classList.add("hidden");
      applyChatContextState();
      return;
    }

    if (isDirectHubEmpty()) {
      q("room-icon").textContent = "DM";
      q("room-title").textContent = "Diretas";
      q("room-description").textContent = "Ainda nao existe DM criada. Escolhe alguem em membros e abre a primeira.";
      q("overview-room-name").textContent = "Diretas";
      q("overview-room-copy").textContent = "Espaco para puxar conversa 1x1 sem ruido.";
      q("room-members-pill").textContent = state.online.filter(function(item) { return item.online; }).length + " on";
      q("stat-version").textContent = "v" + String(state.version || 0);
      q("room-type-badge").className = "badge kind-dm";
      q("room-type-badge").textContent = "dm";
      q("room-access-badge").className = "badge badge-live";
      q("room-access-badge").textContent = "pronta";
      stateCard.classList.add("hidden");
      applyChatContextState();
      return;
    }

    if (!room) {
      q("room-icon").textContent = "PD";
      q("room-title").textContent = "Painel Dief";
      q("room-description").textContent = "Escolhe uma area na esquerda.";
      q("overview-room-name").textContent = "Painel Dief";
      q("overview-room-copy").textContent = "Sem sala ativa.";
      q("room-members-pill").textContent = "--";
      q("stat-version").textContent = "v" + String(state.version || 0);
      q("room-type-badge").className = "badge";
      q("room-type-badge").textContent = "painel";
      q("room-access-badge").className = "badge badge-live";
      q("room-access-badge").textContent = "livre";
      stateCard.classList.add("hidden");
      applyChatContextState();
      return;
    }

    q("room-icon").textContent = displayRoomIcon(room);
    q("room-title").textContent = displayRoomName(room);
    q("room-description").textContent = displayRoomDescription(room);
    q("overview-room-name").textContent = displayRoomName(room);
    q("overview-room-copy").textContent = displayRoomDescription(room);
    memberText = room.scope === "dm" ? "1:1" : activeRoomMembers().length + " on";
    q("room-members-pill").textContent = memberText;
    q("stat-version").textContent = "v" + String(state.version || 0);
    q("room-type-badge").className = "badge kind-" + roomKind(room);
    q("room-type-badge").textContent = roomKindLabel(room);
    q("composer-input").placeholder = state.editingMessage ? "Ajusta o texto e salva a edicao..." : composerPlaceholder(room);
    q("btn-send").textContent = state.sendingMessage
      ? "Enviando..."
      : (hasUploadingPendingUploads()
        ? "Subindo..."
        : (hasErroredPendingUploads()
          ? "Corrige upload"
          : (state.editingMessage ? "Salvar" : "Enviar")));

    if (isAppsLabRoom(room)) {
      q("room-description").textContent = "Sala tecnica privada do admin para build, terminal, codigo e atualizacao do UniversalD.";
      q("overview-room-copy").textContent = "Central privada do UniversalD com abertura, download, hash, build e painel tecnico sem conversa no meio.";
      q("room-type-badge").className = "badge kind-dev";
      q("room-type-badge").textContent = "lab";
      q("composer-form").classList.add("hidden");
      q("message-stream").classList.add("hidden");
      q("typing-indicator").classList.add("hidden");
      q("pins-strip").classList.add("hidden");
      q("unread-banner").classList.add("hidden");
    }

    if (access === "dm") {
      q("room-access-badge").className = "badge kind-dm";
      q("room-access-badge").textContent = "direta";
      stateCard.classList.add("hidden");
      if (peer && (peer.blockedByViewer || peer.hasBlockedViewer)) {
        q("room-access-badge").className = "badge badge-lock";
        q("room-access-badge").textContent = "travada";
        stateCard.classList.remove("hidden");
        q("room-state-title").textContent = peer.blockedByViewer ? "DM bloqueada por ti" : "DM indisponivel";
        q("room-state-copy").textContent = peer.blockedByViewer
          ? "Tu bloqueou esse usuario. Desbloqueia no perfil pra voltar a conversar."
          : "Esse usuario bloqueou teu contato. A DM fica so na leitura.";
        btnAction.textContent = "Abrir perfil";
        composerDisabled = true;
      }
    } else if (access === "locked") {
      q("room-access-badge").className = "badge badge-lock";
      q("room-access-badge").textContent = "senha";
      stateCard.classList.remove("hidden");
      q("room-state-title").textContent = "Sala com senha";
      q("room-state-copy").textContent = "Digita a senha certa ou o painel te deixa no corredor.";
      btnAction.textContent = "Destrancar";
      composerDisabled = true;
    } else if (access === "vip") {
      q("room-access-badge").className = "badge badge-vip";
      q("room-access-badge").textContent = "vip";
      stateCard.classList.remove("hidden");
      q("room-state-title").textContent = "Acesso VIP";
      q("room-state-copy").textContent = "Essa sala e pra VIP, admin ou owner. Hoje ela nao te chamou.";
      btnAction.textContent = "Sem acesso";
      composerDisabled = true;
    } else if (access === "admin") {
      q("room-access-badge").className = "badge badge-admin";
      q("room-access-badge").textContent = "admin";
      stateCard.classList.add("hidden");
    } else {
      q("room-access-badge").className = "badge badge-live";
      q("room-access-badge").textContent = "aberta";
      stateCard.classList.add("hidden");
    }

    q("composer-input").disabled = composerDisabled || state.sendingMessage;
    q("btn-send").disabled = composerDisabled || state.sendingMessage || hasUploadingPendingUploads();
    q("btn-attach").disabled = composerDisabled || state.sendingMessage;
    q("btn-ai").classList.toggle("hidden", room.slug !== "nego-dramias-ia");
    q("btn-ai").disabled = composerDisabled || state.sendingMessage || hasUploadingPendingUploads();
    syncComposerDraftHint();
    applyChatContextState();
  }

  function renderDashboard() {
    var room = activeRoom();
    var attachments = currentRoomAttachments();
    var mode = accessForRoom(room);
    var onlineNow = state.online.filter(function(item) { return item.online; }).length;
    var quickline = q("dashboard-quickline");
    q("dashboard-online-now").textContent = String(onlineNow);
    q("dashboard-theme").textContent = themeLabel(state.viewer && state.viewer.theme || "matrix");
    if (isSusView()) {
      q("dashboard-last-activity").textContent = Number(state.susClicks || 0) ? ("clique " + Number(state.susClicks || 0)) : "silencio ruim";
      q("dashboard-room-files").textContent = "0";
      q("dashboard-room-mode").textContent = "sus // anomalia";
      q("dashboard-nego-tip").textContent = negoDashboardTip(room, mode);
      quickline.innerHTML = dashboardQuicklineItems(room, mode, attachments).map(function(item) {
        return "<span class='ghost-pill ghost-pill-soft'>" + esc(item) + "</span>";
      }).join("");
      return;
    }
    if (isMembersHub()) {
      q("dashboard-last-activity").textContent = onlineNow ? (onlineNow + " ativos") : "base vazia";
      q("dashboard-room-files").textContent = String(state.online.length);
      q("dashboard-room-mode").textContent = roomModeLabel(room, mode);
      q("dashboard-nego-tip").textContent = negoDashboardTip(room, mode);
      quickline.innerHTML = dashboardQuicklineItems(room, mode, attachments).map(function(item) {
        return "<span class='ghost-pill ghost-pill-soft'>" + esc(item) + "</span>";
      }).join("");
      return;
    }
    if (isDirectHubEmpty()) {
      q("dashboard-last-activity").textContent = "sem DM aberta";
      q("dashboard-room-files").textContent = String(state.rooms.filter(function(item) { return item.scope === "dm"; }).length);
      q("dashboard-room-mode").textContent = roomModeLabel(room, mode);
      q("dashboard-nego-tip").textContent = negoDashboardTip(room, mode);
      quickline.innerHTML = dashboardQuicklineItems(room, mode, attachments).map(function(item) {
        return "<span class='ghost-pill ghost-pill-soft'>" + esc(item) + "</span>";
      }).join("");
      return;
    }
    if (!room) {
      q("dashboard-last-activity").textContent = "--";
      q("dashboard-room-files").textContent = "0";
      q("dashboard-room-mode").textContent = roomModeLabel(room, mode);
      q("dashboard-nego-tip").textContent = negoDashboardTip(room, mode);
      quickline.innerHTML = "";
      return;
    }
    q("dashboard-last-activity").textContent = room.lastMessageAt ? formatDateTime(room.lastMessageAt) : "sem atividade";
    q("dashboard-room-files").textContent = String(attachments.length);
    q("dashboard-room-mode").textContent = isAppsLabRoom(room) ? "lab privado // admin" : (roomModeLabel(room, mode) + " // " + roomPulseLabel(room));
    q("dashboard-nego-tip").textContent = isAppsLabRoom(room)
      ? "Bah, aqui o papo eh tecnico: admins trocam build, terminal, codigo e update do UniversalD sem poluir o resto da base."
      : negoDashboardTip(room, mode);
    quickline.innerHTML = dashboardQuicklineItems(room, mode, attachments).map(function(item) {
      return "<span class='ghost-pill ghost-pill-soft'>" + esc(item) + "</span>";
    }).join("");
  }

  function renderAppsLabDeck() {
    var room = activeRoom();
    var visible = !!(isAppsLabRoom(room) && accessForRoom(room) === "admin");
    var wrap = q("apps-lab-deck");
    if (!wrap) {
      return;
    }
    wrap.classList.toggle("hidden", !visible);
    if (!visible) {
      return;
    }
    q("apps-app-version-pill").textContent = UNIVERSALD_META.shortVersion;
    q("apps-build-copy").textContent = UNIVERSALD_META.version;
    q("apps-size-copy").textContent = UNIVERSALD_META.sizeLabel;
    q("apps-sha-copy").textContent = UNIVERSALD_META.sha256.slice(0, 12) + "...";
  }

  function renderSusPanel() {
    var wrap = q("sus-panel");
    var position = currentSusPosition();
    var warning = currentSusWarning();
    if (!wrap) {
      return;
    }
    if (!isSusView()) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
      wrap.innerHTML =
        "<div class='sus-shell'>" +
          "<span class='sus-eyebrow'>aba suspeita</span>" +
          "<h3>Tem alguma coisa errada aqui.</h3>" +
          "<p class='sus-copy'>Essa area nao conversa, nao explica e nao ajuda. Ela so provoca, treme e te encara de volta.</p>" +
          "<div class='sus-warning'>" + esc(warning) + "</div>" +
          "<div class='sus-stage'>" +
            "<button type='button' class='sus-trigger' data-action='sus-trigger' style='left: calc(50% + " + Number(position.x || 0) + "px); top: calc(50% + " + Number(position.y || 0) + "px);'>abrir mesmo assim?</button>" +
          "</div>" +
        "<div class='inline-row'>" +
          "<span class='ghost-pill'>cliques: " + Number(state.susClicks || 0) + "/4</span>" +
          "<span class='ghost-pill'>status: anomalia</span>" +
          "<span class='ghost-pill'>destino: desconhecido</span>" +
        "</div>" +
      "</div>";
  }

  function handleSusTrigger() {
    state.susClicks = Number(state.susClicks || 0) + 1;
    state.susPositionIndex = Math.min(SUS_POSITIONS.length - 1, Number(state.susPositionIndex || 0) + 1);
    playTone("sus");
    showSusOmen(currentSusWarning());
    renderSusPanel();
    renderDashboard();
    if (state.susClicks >= 4) {
      toast("Bah... agora tu vai ter que ver o que tem do outro lado.", "warn");
      window.setTimeout(function() {
        window.location.assign(SUS_TARGET_URL);
      }, 900);
      return;
    }
    toast(currentSusWarning(), "warn");
  }

  function renderPendingAttachment() {
    var wrap = q("pending-attachment");
    var items = pendingUploads();
    syncPendingAttachmentAlias();
    if (!items.length) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.innerHTML = items.map(function(item) {
      var label = item.attachment
        ? (item.attachment.name + " // " + item.attachment.kind + " // " + bytesLabel(item.attachment.sizeBytes))
        : (item.fileName + " // " + bytesLabel(item.sizeBytes || 0));
      var status = item.status === "uploading"
        ? ("Subindo " + Math.max(1, Math.min(100, Number(item.progress || 1))) + "%")
        : (item.status === "error" ? ("Falhou // " + (item.error || "erro no upload")) : "Pronto");
      return "<div class='pending-upload-row " + esc(item.status || "ready") + "'>" +
        "<span><strong>" + esc(label) + "</strong><small>" + esc(status) + "</small></span>" +
        "<div class='pending-upload-actions'>" +
          (item.status === "error" ? "<button type='button' class='btn btn-ghost' data-action='retry-upload' data-upload-id='" + esc(item.id) + "'>Tentar de novo</button>" : "") +
          "<button type='button' class='btn btn-ghost' data-action='remove-upload' data-upload-id='" + esc(item.id) + "'>Remover</button>" +
        "</div>" +
      "</div>";
    }).join("");
  }

  function removePendingUpload(uploadId) {
    state.pendingUploads = pendingUploads().filter(function(item) {
      return String(item.id) !== String(uploadId);
    });
    syncPendingAttachmentAlias();
    renderPendingAttachment();
  }

  function updatePendingUpload(uploadId, patch) {
    state.pendingUploads = pendingUploads().map(function(item) {
      if (String(item.id) !== String(uploadId)) {
        return item;
      }
      var next = {};
      Object.keys(item).forEach(function(key) {
        next[key] = item[key];
      });
      Object.keys(patch || {}).forEach(function(key) {
        next[key] = patch[key];
      });
      return next;
    });
    syncPendingAttachmentAlias();
    renderPendingAttachment();
  }

  function clearErroredPendingUploads() {
    state.pendingUploads = pendingUploads().filter(function(item) {
      return item && item.status !== "error";
    });
    syncPendingAttachmentAlias();
    renderPendingAttachment();
  }

  function createPendingUpload(file) {
    return {
      id: "upl-" + Date.now() + "-" + Math.random().toString(16).slice(2),
      file: file,
      fileName: file.name,
      sizeBytes: file.size,
      progress: 0,
      status: "uploading",
      error: "",
      attachment: null
    };
  }

  function uploadWithProgress(file, onProgress) {
    return new Promise(function(resolve, reject) {
      var formData = new FormData();
      var request = new XMLHttpRequest();
      var storedSessionId = readStoredSessionId();
      formData.append("file", file);
      request.open("POST", "/api/panel/upload", true);
      request.withCredentials = true;
      request.timeout = 45000;
      if (storedSessionId) {
        try {
          request.setRequestHeader("X-Panel-Session", storedSessionId);
        } catch (headerErr) {}
      }
      request.upload.onprogress = function(event) {
        if (!event.lengthComputable || !onProgress) {
          return;
        }
        onProgress(Math.round((event.loaded / event.total) * 100));
      };
      request.onreadystatechange = function() {
        var payload = {};
        if (request.readyState !== 4) {
          return;
        }
        try {
          payload = JSON.parse(request.responseText || "{}");
        } catch (e) {}
        if (request.status >= 200 && request.status < 300) {
          resolve(payload.attachments || (payload.attachment ? [payload.attachment] : []));
          return;
        }
        if (request.status === 401) {
          resetSession();
        }
        reject(new Error((payload && payload.error) || "falha inesperada no upload"));
      };
      request.onerror = function() {
        reject(new Error("rede falhou no meio do upload"));
      };
      request.ontimeout = function() {
        reject(new Error("upload demorou demais e a rede cansou"));
      };
      request.send(formData);
    });
  }

  async function queueFilesForUpload(fileList, options) {
    var files = Array.prototype.slice.call(fileList || []);
    var mode = options && options.mode || "composer";
    if (!files.length) {
      return [];
    }
    var uploaded = [];
    for (var i = 0; i < files.length; i++) {
      var entry = createPendingUpload(files[i]);
      if (mode === "composer") {
        state.pendingUploads = pendingUploads().concat([entry]);
        renderPendingAttachment();
      }
      try {
        var attachments = await uploadWithProgress(files[i], function(progress) {
          if (mode === "composer") {
            updatePendingUpload(entry.id, { progress: progress, status: "uploading" });
          }
        });
        if (!attachments.length) {
          throw new Error("o servidor nao devolveu o anexo pronto");
        }
        if (mode === "composer") {
          updatePendingUpload(entry.id, {
            progress: 100,
            status: "ready",
            attachment: attachments[0],
            error: ""
          });
        }
        uploaded = uploaded.concat(attachments);
      } catch (err) {
        if (mode === "composer") {
          updatePendingUpload(entry.id, {
            status: "error",
            error: err.message || "falha no upload"
          });
          toast(err.message || "falha no upload", "err");
        } else {
          throw err;
        }
      }
    }
    syncPendingAttachmentAlias();
    return uploaded;
  }

  async function retryPendingUpload(uploadId) {
    var item = pendingUploads().find(function(entry) {
      return String(entry.id) === String(uploadId);
    });
    if (!item || !item.file) {
      return;
    }
    updatePendingUpload(item.id, { status: "uploading", error: "", progress: 0, attachment: null });
    try {
      var attachments = await uploadWithProgress(item.file, function(progress) {
        updatePendingUpload(item.id, { progress: progress, status: "uploading" });
      });
      if (!attachments.length) {
        throw new Error("o servidor nao devolveu o anexo pronto");
      }
      updatePendingUpload(item.id, {
        progress: 100,
        status: "ready",
        attachment: attachments[0],
        error: ""
      });
      toast("Upload retomado com sucesso.", "ok");
    } catch (err) {
      updatePendingUpload(item.id, {
        status: "error",
        error: err.message || "falha no retry"
      });
      toast(err.message || "falha no retry", "err");
    }
  }

  function findMessageInRoom(roomId, messageId) {
    var items = state.messagesByRoom[String(roomId)] || [];
    for (var i = 0; i < items.length; i++) {
      if (Number(items[i].id) === Number(messageId)) {
        return items[i];
      }
    }
    return null;
  }

  function currentPreviewIndex() {
    var attachments = roomAttachmentsForPreview(state.mediaPreview && state.mediaPreview.roomId);
    var i;
    if (!state.mediaPreview) {
      return -1;
    }
    for (i = 0; i < attachments.length; i++) {
      if (Number(attachments[i].id) === Number(state.mediaPreview.messageId)) {
        return i;
      }
    }
    return -1;
  }

  function renderMediaPreview() {
    var modal = q("media-modal");
    var stage = q("media-modal-stage");
    var meta = q("media-modal-meta");
    var filmstrip = q("media-modal-filmstrip");
    var originButton = q("media-modal-origin");
    var current;
    var attachment;
    var sameRoomAttachments;
    var index;
    var room;
    var excerpt;
    var attachmentType;
    if (!state.mediaPreview) {
      modal.classList.add("hidden");
      stage.innerHTML = "";
      meta.innerHTML = "";
      if (filmstrip) {
        filmstrip.innerHTML = "";
      }
      if (originButton) {
        originButton.disabled = true;
        originButton.removeAttribute("data-room-id");
        originButton.removeAttribute("data-message-id");
      }
      return;
    }
    current = findMessageInRoom(state.mediaPreview.roomId, state.mediaPreview.messageId);
    if (!current || !current.attachment) {
      state.mediaPreview = null;
      modal.classList.add("hidden");
      return;
    }
    attachment = current.attachment;
    var attachmentUrl = safeAttachmentUrl(attachment.url);
    sameRoomAttachments = roomAttachmentsForPreview(current.roomId);
    index = currentPreviewIndex();
    room = roomById(current.roomId);
    excerpt = String(current.body || "").trim();
    attachmentType = attachment.kind === "image"
      ? "Imagem"
      : (attachment.kind === "video"
        ? "Video"
        : (attachment.kind === "audio"
          ? "Audio"
          : "Arquivo"));
    q("media-modal-title").textContent = attachment.name || "Preview de midia";
    if (attachmentUrl && attachment.kind === "image") {
      stage.innerHTML = "<img alt='" + esc(attachment.name) + "' src='" + esc(attachmentUrl) + "' />";
    } else if (attachmentUrl && attachment.kind === "video") {
      stage.innerHTML = "<video controls autoplay src='" + esc(attachmentUrl) + "'></video>";
    } else if (attachmentUrl && attachment.kind === "audio") {
      stage.innerHTML = "<audio controls autoplay src='" + esc(attachmentUrl) + "'></audio>";
    } else if (attachmentUrl && attachment.contentType === "application/pdf") {
      stage.innerHTML = "<iframe title='" + esc(attachment.name) + "' src='" + esc(attachmentUrl) + "'></iframe>";
    } else {
      stage.innerHTML = "<div class='media-file-fallback'><strong>" + esc(attachment.name) + "</strong><span>" + esc(attachment.contentType || "arquivo") + "</span></div>";
    }
    meta.innerHTML =
      "<article class='info-card'>" +
        "<span>Tipo</span><strong>" + esc(attachmentType) + "</strong>" +
        "<p>" + esc(attachment.contentType || "sem content-type") + "</p>" +
      "</article>" +
      "<article class='info-card'>" +
        "<span>Tamanho</span><strong>" + esc(bytesLabel(attachment.sizeBytes)) + "</strong>" +
        "<p>" + esc(current.authorName || "autor desconhecido") + " // " + esc(formatDateTime(current.createdAt)) + "</p>" +
      "</article>" +
      "<article class='info-card media-origin-card'>" +
        "<span>Sala</span><strong>" + esc(room ? displayRoomName(room) : "Sem sala") + "</strong>" +
        "<p>" + esc(excerpt ? excerpt.slice(0, 160) : "Sem legenda na mensagem de origem.") + "</p>" +
        "<div class='inline-row media-origin-pills'>" +
          "<span class='ghost-pill'>" + esc(current.authorName || "autor desconhecido") + "</span>" +
          "<span class='ghost-pill'>" + esc(formatRelative(current.createdAt)) + "</span>" +
        "</div>" +
      "</article>" +
      ((attachment.width || attachment.height) ? ("<article class='info-card'><span>Resolucao</span><strong>" + esc((attachment.width || 0) + " x " + (attachment.height || 0)) + "</strong><p>" + esc(attachment.extension || "") + "</p></article>") : "") +
      "<article class='info-card media-origin-card soft'>" +
        "<span>Fluxo</span><strong>" + esc(index + 1) + " / " + esc(sameRoomAttachments.length || 1) + "</strong>" +
        "<p>" + esc(room && room.summary ? room.summary : "Preview vivo do historico dessa sala.") + "</p>" +
        "<div class='inline-row media-origin-pills'>" +
          "<span class='ghost-pill'>" + esc(attachment.kind || "arquivo") + "</span>" +
          "<span class='ghost-pill'>" + esc(bytesLabel(attachment.sizeBytes)) + "</span>" +
        "</div>" +
      "</article>";
    if (filmstrip) {
      var windowed = sameRoomAttachments.slice(Math.max(0, index - 2), Math.min(sameRoomAttachments.length, index + 4));
      filmstrip.innerHTML = windowed.map(function(item) {
        var itemAttachment = item.attachment || {};
        var itemUrl = safeAttachmentUrl(itemAttachment.url);
        var active = Number(item.id) === Number(current.id);
        var thumb = itemUrl && itemAttachment.kind === "image"
          ? "<img alt='" + esc(itemAttachment.name || "midia") + "' src='" + esc(itemUrl) + "' />"
          : (itemUrl && itemAttachment.kind === "video"
            ? "<video preload='metadata' muted src='" + esc(itemUrl) + "'></video>"
            : "<div class='media-film-fallback'>" + esc((itemAttachment.kind || "arquivo").slice(0, 6)) + "</div>");
        return "<button type='button' class='media-film-item" + (active ? " active" : "") + "' data-action='preview-attachment' data-room-id='" + Number(item.roomId) + "' data-message-id='" + Number(item.id) + "'>" +
          thumb +
          "<span class='media-film-label'>" + esc(itemAttachment.kind || "arquivo") + "</span>" +
        "</button>";
      }).join("");
    }
    q("media-modal-download").href = attachmentUrl || "#";
    q("media-modal-prev").disabled = index <= 0;
    q("media-modal-next").disabled = index < 0 || index >= (sameRoomAttachments.length - 1);
    if (originButton) {
      originButton.disabled = false;
      originButton.setAttribute("data-room-id", Number(current.roomId));
      originButton.setAttribute("data-message-id", Number(current.id));
    }
    modal.classList.remove("hidden");
    syncTransientLayoutState();
  }

  function openMediaPreview(roomId, messageId) {
    if (state.isMobile) {
      closeSidebar();
      closeInspector();
      state.utilityStripCollapsed = true;
      syncUtilityStripUI();
    }
    state.mediaPreview = { roomId: Number(roomId), messageId: Number(messageId) };
    renderMediaPreview();
  }

  function stepMediaPreview(delta) {
    var items = roomAttachmentsForPreview(state.mediaPreview && state.mediaPreview.roomId);
    var index = currentPreviewIndex();
    var next;
    if (index < 0) {
      return;
    }
    next = items[index + delta];
    if (!next) {
      return;
    }
    state.mediaPreview = { roomId: Number(next.roomId), messageId: Number(next.id) };
    renderMediaPreview();
  }

  function renderReplyChip() {
    var wrap = q("reply-chip");
    if (!state.replyTarget) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.innerHTML =
      "<span>Respondendo " + esc(state.replyTarget.authorName || "mensagem") + ": " + esc((state.replyTarget.body || "").slice(0, 110) || "[anexo]") + "</span>" +
      "<button type='button' id='btn-clear-reply' class='btn btn-ghost'>Cancelar</button>";
  }

  function renderEditChip() {
    var wrap = q("edit-chip");
    if (!state.editingMessage) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.innerHTML =
      "<span>Editando mensagem de " + esc(state.editingMessage.authorName || "alguem") + ". Salva sem enterrar o contexto.</span>" +
      "<button type='button' id='btn-clear-edit' class='btn btn-ghost'>Cancelar edicao</button>";
  }

  function renderPinnedStrip() {
    var wrap = q("pins-strip");
    var items = currentRoomPins();
    if (isHubView() || isAppsLabRoom(activeRoom()) || !items.length) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.innerHTML =
      "<div class='pins-head'><strong>Fixados</strong><span>" + items.length + " no topo</span></div>" +
      items.map(function(message) {
        var preview = message.body || (message.attachment ? "[anexo] " + message.attachment.name : "[sem texto]");
        return "<button type='button' class='pin-card' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>" +
          "<strong>" + esc(message.authorName || "Mensagem") + "</strong>" +
          "<span>" + esc(preview.slice(0, 140)) + "</span>" +
        "</button>";
      }).join("");
  }

  function renderFavorites() {
    var wrap = q("favorite-list");
    var items = currentRoomMessages().filter(function(message) {
      return !!message.viewerFavorited;
    });
    q("favorites-status-pill").textContent = items.length + " itens";
    if (isHubView()) {
      wrap.innerHTML = "<div class='empty-state'>Abre um chat para ver os favoritos dessa sala.</div>";
      return;
    }
    if (!items.length) {
      wrap.innerHTML = "<div class='empty-state'>Favorita mensagens e elas aparecem aqui pra achar rapido.</div>";
      return;
    }
    wrap.innerHTML = items.map(function(message) {
      var preview = message.body || (message.attachment ? "[anexo] " + message.attachment.name : "[sem texto]");
      return "<article class='favorite-card'>" +
        "<strong>" + esc(message.authorName || "Mensagem") + "</strong>" +
        "<p>" + esc(preview.slice(0, 160)) + "</p>" +
        "<button class='btn btn-ghost' type='button' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Abrir</button>" +
      "</article>";
    }).join("");
  }

  function messageAvatarMarkup(message) {
    var isViewer = state.viewer && Number(message.authorId) === Number(state.viewer.id);
    var viewerAvatarUrl = safeAvatarUrl(state.viewer && state.viewer.avatarUrl);
    var avatarOpen =
      "<button type='button' class='identity-trigger avatar-trigger' data-action='open-user-profile' data-user-id='" + Number(message.authorId) + "' aria-label='Abrir perfil de " + esc(message.authorName) + "'>";
    if (isViewer && viewerAvatarUrl) {
      return avatarOpen + "<img class='message-avatar' alt='" + esc(message.authorName) + "' src='" + esc(viewerAvatarUrl) + "' /></button>";
    }
    var profile = presenceByUserId(message.authorId);
    var profileAvatarUrl = safeAvatarUrl(profile && profile.avatarUrl);
    if (profileAvatarUrl) {
      return avatarOpen + "<img class='message-avatar' alt='" + esc(message.authorName) + "' src='" + esc(profileAvatarUrl) + "' /></button>";
    }
    return avatarOpen + "<div class='message-avatar'>" + esc(initials(message.authorName || "PD")) + "</div></button>";
  }

  function attachmentMarkup(message) {
    var attachment = message && message.attachment;
    var attachmentUrl = safeAttachmentUrl(attachment && attachment.url);
    var previewAttrs = " data-action='preview-attachment' data-room-id='" + Number(message && message.roomId || 0) + "' data-message-id='" + Number(message && message.id || 0) + "'";
    if (!attachment) {
      return "";
    }
    if (attachmentUrl && attachment.kind === "image") {
      return "<button type='button' class='message-attachment attachment-preview' " + previewAttrs + ">" +
        "<img alt='" + esc(attachment.name) + "' src='" + esc(attachmentUrl) + "' />" +
        "<span class='attachment-caption'>" + esc(attachment.name) + "</span>" +
      "</button>";
    }
    if (attachmentUrl && attachment.kind === "video") {
      return "<button type='button' class='message-attachment attachment-preview' " + previewAttrs + ">" +
        "<video preload='metadata' muted src='" + esc(attachmentUrl) + "'></video>" +
        "<span class='attachment-caption'>Video // " + esc(attachment.name) + "</span>" +
      "</button>";
    }
    if (attachmentUrl && attachment.kind === "audio") {
      return "<div class='message-attachment'>" +
        "<audio controls preload='metadata' src='" + esc(attachmentUrl) + "'></audio>" +
        "<div class='attachment-actions'>" +
          "<button type='button' class='btn btn-ghost' " + previewAttrs + ">Abrir</button>" +
          "<a class='btn btn-ghost' href='" + esc(attachmentUrl) + "' target='_blank' rel='noreferrer'>Baixar</a>" +
        "</div>" +
      "</div>";
    }
    return "<div class='message-attachment file-attachment'>" +
      "<div class='file-attachment-copy'>" +
        "<strong>" + esc(attachment.name || "arquivo") + "</strong>" +
        "<span class='media-meta'>" + esc((attachment.extension || attachment.kind || "arquivo") + " // " + bytesLabel(attachment.sizeBytes)) + "</span>" +
      "</div>" +
      "<div class='attachment-actions'>" +
        "<button type='button' class='btn btn-ghost' " + previewAttrs + ">Abrir</button>" +
        (attachmentUrl ? "<a class='btn btn-ghost' href='" + esc(attachmentUrl) + "' target='_blank' rel='noreferrer'>Baixar</a>" : "") +
      "</div>" +
    "</div>";
  }

  function replyMarkup(reply) {
    if (!reply) {
      return "";
    }
    return "<div class='message-reply'><strong>" + esc(reply.authorName || "Mensagem") + "</strong><span>" + esc(reply.bodyPreview || "[sem texto]") + "</span></div>";
  }

  function messageStatusMarkup(message) {
    var badges = [];
    if (message.blockedByViewer) {
      badges.push("<span class='badge'>bloqueado</span>");
    }
    if (message.isPinned) {
      badges.push("<span class='badge badge-admin'>fixada</span>");
    }
    if (message.viewerFavorited) {
      badges.push("<span class='badge badge-vip'>favorita</span>");
    }
    if (message.updatedAt) {
      badges.push("<span class='badge'>editada</span>");
    }
    if (messageMentionsViewer(message) && Number(message.authorId) !== Number(state.viewer && state.viewer.id)) {
      badges.push("<span class='badge badge-live'>te marcou</span>");
    }
    if (!badges.length) {
      return "";
    }
    return "<div class='message-status-row'>" + badges.join("") + "</div>";
  }

  function messageActionsMarkup(message) {
    if (message.blockedByViewer) {
      return "<button type='button' class='btn btn-ghost' data-action='open-user-profile' data-user-id='" + Number(message.authorId) + "'>Ver perfil</button>";
    }
    var buttons = [
      "<button type='button' class='btn btn-ghost' data-action='reply' data-message-id='" + Number(message.id) + "'>Responder</button>",
      "<button type='button' class='btn btn-ghost' data-action='favorite' data-message-id='" + Number(message.id) + "'>" + (message.viewerFavorited ? "Desfavoritar" : "Favoritar") + "</button>",
      "<button type='button' class='btn btn-ghost' data-action='copy' data-message-id='" + Number(message.id) + "'>Copiar</button>"
    ];
    if (canPinMessage(message)) {
      buttons.splice(2, 0, "<button type='button' class='btn btn-ghost' data-action='pin' data-message-id='" + Number(message.id) + "'>" + (message.isPinned ? "Desfixar" : "Fixar") + "</button>");
    }
    if (canManageMessage(message)) {
      buttons.push("<button type='button' class='btn btn-ghost' data-action='edit' data-message-id='" + Number(message.id) + "'>Editar</button>");
      buttons.push("<button type='button' class='btn btn-ghost' data-action='delete' data-message-id='" + Number(message.id) + "'>Apagar</button>");
    }
    return buttons.join("");
  }

  function reactionsMarkup(message) {
    var existing = (message.reactions || []).map(function(reaction) {
      return "<button type='button' class='reaction-chip" + (reaction.viewerReacted ? " active" : "") + "' data-action='react' data-message-id='" + Number(message.id) + "' data-emoji='" + esc(reaction.emoji) + "'>" + esc(reaction.emoji) + " " + reaction.count + "</button>";
    }).join("");
    var quick = QUICK_REACTIONS.map(function(emoji) {
      return "<button type='button' class='quick-reaction' data-action='react' data-message-id='" + Number(message.id) + "' data-emoji='" + esc(emoji) + "'>" + esc(emoji) + "</button>";
    }).join("");
    return "<div class='message-reactions'>" + existing + "</div><div class='quick-reactions'>" + quick + "</div>";
  }

  function shouldCondenseMessage(previous, current) {
    var previousAt = 0;
    var currentAt = 0;
    if (!previous || !current) {
      return false;
    }
    if (Number(previous.authorId) !== Number(current.authorId)) {
      return false;
    }
    if (!!previous.isAI !== !!current.isAI) {
      return false;
    }
    if (current.reply) {
      return false;
    }
    previousAt = new Date(previous.createdAt).getTime();
    currentAt = new Date(current.createdAt).getTime();
    if (!previousAt || !currentAt) {
      return false;
    }
    return (currentAt - previousAt) < (5 * 60 * 1000);
  }

  function renderMessages(roomId) {
    var stream = q("message-stream");
    var room = roomById(roomId);
    var items = state.messagesByRoom[String(roomId)] || [];
    var highlightNode = null;
    var onlineItems = state.online.filter(function(item) { return item.online; });
    stream.innerHTML = "";
    stream.classList.remove("hidden");
    if (isSusView()) {
      stream.classList.add("hidden");
      renderSusPanel();
      renderTypingIndicator();
      return;
    }
    if (isMembersHub() || isDirectHubEmpty()) {
      if (!onlineItems.length) {
        stream.innerHTML = "<div class='empty-state'>Ninguem esta online agora. Quando o grupo voltar, esse hub enche sozinho.</div>";
        renderTypingIndicator();
        return;
      }
      onlineItems.forEach(function(person) {
        var canDM = state.viewer && Number(person.userId) !== Number(state.viewer.id) && person.role !== "ai";
        var avatarUrl = safeAvatarUrl(person.avatarUrl);
        var statusCopy = personStatusCopy(person);
        var card = document.createElement("article");
        card.className = "message-card";
        card.innerHTML =
          "<button type='button' class='identity-trigger avatar-trigger' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>" +
            (avatarUrl ? "<img class='message-avatar' alt='" + esc(person.displayName || person.username) + "' src='" + esc(avatarUrl) + "' />" : "<div class='message-avatar'>" + esc(initials(person.displayName || person.username)) + "</div>") +
          "</button>" +
          "<div class='message-wrap'>" +
            "<div class='message-head'>" +
              "<button type='button' class='identity-trigger identity-name' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>" + esc(person.displayName || person.username) + "</button>" +
              "<span class='message-role'>" + esc(person.role || "member") + "</span>" +
              "<span class='message-time'>" + esc(person.status || "online") + "</span>" +
            "</div>" +
            "<p class='message-body'>" + esc(statusCopy || (canDM ? "Abre uma DM direta com esse vivente e conversa sem ruido." : "Presenca carregada no grupo.")) + "</p>" +
            "<div class='message-actions'>" +
              "<button type='button' class='btn btn-ghost' data-action='open-user-profile' data-user-id='" + Number(person.userId) + "'>Perfil</button>" +
              (canDM ? "<button type='button' class='btn btn-ghost' data-action='open-dm' data-user-id='" + Number(person.userId) + "'>Abrir DM</button>" : "") +
            "</div>" +
          "</div>";
        stream.appendChild(card);
      });
      renderTypingIndicator();
      return;
    }
    if (!room) {
      stream.innerHTML = "<div class='empty-state'>Escolhe uma area na esquerda pra abrir o chat certo.</div>";
      return;
    }
    if (isAppsLabRoom(room)) {
      stream.classList.add("hidden");
      renderTypingIndicator();
      return;
    }
    if (accessForRoom(room) === "locked" || accessForRoom(room) === "vip") {
      stream.innerHTML = "<div class='empty-state'>Essa aba ainda nao abriu de verdade.</div>";
      renderTypingIndicator();
      return;
    }
    if (!items.length) {
      stream.innerHTML = "<div class='empty-state'>A sala ta silenciosa. Manda a primeira mensagem.</div>";
      renderTypingIndicator();
      return;
    }
    items.forEach(function(message, index) {
      var card = document.createElement("article");
      var bodyMarkup = message.blockedByViewer
        ? "<p class='message-body blocked-copy'>Mensagem escondida porque esse usuario esta bloqueado pra ti.</p>"
        : renderMessageBodyMarkup(message);
      var compact = shouldCondenseMessage(items[index - 1], message);
      var highlight = Number(state.highlightMessageId) === Number(message.id) ? " highlight" : "";
      var mentionClass = messageMentionsViewer(message) && Number(message.authorId) !== Number(state.viewer && state.viewer.id) ? " mention-hit" : "";
      var headMarkup = compact
        ? "<div class='message-head compact-head'><span class='message-time'>" + esc(formatTime(message.createdAt)) + "</span></div>"
        : "<div class='message-head'>" +
            "<button type='button' class='identity-trigger identity-name' data-action='open-user-profile' data-user-id='" + Number(message.authorId) + "'>" + esc(message.authorName || "Anonimo") + "</button>" +
            "<span class='message-role'>" + esc(message.authorRole || "member") + "</span>" +
            "<span class='message-time'>" + esc(formatDateTime(message.createdAt)) + "</span>" +
          "</div>";
      card.className = "message-card" + (message.isAI ? " ai" : "") + (compact ? " compact" : "") + highlight + mentionClass;
      if (highlight) {
        highlightNode = card;
      }
      card.innerHTML =
        (compact ? "<div class='message-avatar compact-spacer' aria-hidden='true'></div>" : messageAvatarMarkup(message)) +
        "<div class='message-wrap'>" +
          headMarkup +
          messageStatusMarkup(message) +
          (message.blockedByViewer ? "" : replyMarkup(message.reply)) +
          bodyMarkup +
          (message.blockedByViewer ? "" : attachmentMarkup(message)) +
          (message.blockedByViewer ? "" : reactionsMarkup(message)) +
          "<div class='message-actions'>" +
            messageActionsMarkup(message) +
          "</div>" +
        "</div>";
      stream.appendChild(card);
    });
    if (highlightNode) {
      safeScrollIntoView(highlightNode, { block: "center", behavior: "smooth" });
    } else {
      stream.scrollTop = stream.scrollHeight;
    }
    renderTypingIndicator();
  }

  function renderTypingIndicator() {
    var wrap = q("typing-indicator");
    if (isSusView() || isAppsLabRoom(activeRoom())) {
      wrap.classList.add("hidden");
      wrap.textContent = "";
      return;
    }
    if (isHubView()) {
      wrap.classList.add("hidden");
      wrap.textContent = "";
      return;
    }
    var active = state.typing.filter(function(item) {
      return Number(item.roomId) === Number(state.activeRoomId);
    });
    if (!active.length) {
      wrap.classList.add("hidden");
      wrap.textContent = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.textContent = active.map(function(item) { return item.displayName; }).join(", ") + " esta digitando...";
  }

  function renderActivity() {
    var feed = q("activity-feed");
    var items = activityFeedItems(8);
    feed.innerHTML = "";
    if (!items.length) {
      feed.innerHTML = "<div class='empty-state'>As notificacoes vao aparecer aqui.</div>";
      syncNotificationCenterButton();
      return;
    }
    items.forEach(function(item) {
      var card = document.createElement("article");
      card.className = "activity-item " + (item.tone || "ok");
      card.innerHTML =
        "<strong>" + esc(formatTime(item.at)) + "</strong>" +
        "<span>" + esc(item.text) + "</span>";
      feed.appendChild(card);
    });
    syncNotificationCenterButton();
  }

  function renderEventRoomOptions() {
    var select = q("event-room-id");
    var value = String(select.value || "0");
    var rooms = (state.rooms || []).filter(function(room) {
      var access = accessForRoom(room);
      return access === "open" || access === "admin" || access === "dm";
    });
    select.innerHTML = "<option value='0'>Evento geral do grupo</option>" + rooms.map(function(room) {
      return "<option value='" + Number(room.id) + "'>" + esc(displayRoomName(room)) + "</option>";
    }).join("");
    if (select.querySelector("option[value='" + value + "']")) {
      select.value = value;
    }
    if (!q("event-starts-at").value) {
      var seed = new Date(Date.now() + 60 * 60 * 1000);
      q("event-starts-at").value = seed.toISOString().slice(0, 16);
    }
  }

  function renderEvents() {
    var list = q("event-list");
    var form = q("event-form");
    var room = activeRoom();
    var items = (state.events || []).slice(0).sort(function(a, b) {
      return new Date(a.startsAt).getTime() - new Date(b.startsAt).getTime();
    });
    q("events-status-pill").textContent = items.length + " eventos";
    renderEventRoomOptions();
    form.classList.toggle("hidden", !state.viewer || state.viewer.role === "ai");
    if (isAppsLabRoom(room)) {
      form.classList.add("hidden");
      list.innerHTML = "<div class='empty-state'>Apps Lab agora e central do app. Agenda ficou fora dessa area.</div>";
      return;
    }
    if (!items.length) {
      list.innerHTML = "<div class='empty-state'>Ainda nao tem evento marcado. Cria um rolê e movimenta a base.</div>";
      return;
    }
    list.innerHTML = items.map(function(event) {
      var roomLabel = event.roomId ? (" // " + (event.roomName || "sala")) : "";
      var manageAction = canManageOwnedContent(event.createdBy)
        ? "<button class='btn btn-ghost btn-danger' type='button' data-action='delete-event' data-event-id='" + Number(event.id) + "'>Apagar</button>"
        : "";
      return "<article class='event-card'>" +
        "<div class='event-copy'>" +
          "<strong>" + esc(event.title) + "</strong>" +
          "<span>" + esc(formatDateTime(event.startsAt) + roomLabel) + "</span>" +
          "<p>" + esc(event.description || "Sem descricao extra.") + "</p>" +
        "</div>" +
        "<div class='event-actions'>" +
          "<span class='ghost-pill'>" + esc(event.rsvpCount + " confirmados") + "</span>" +
          "<div class='inline-row'>" +
            "<button class='btn btn-ghost' type='button' data-action='toggle-event-rsvp' data-event-id='" + Number(event.id) + "'>" + (event.viewerJoined ? "Sair da lista" : "Confirmar") + "</button>" +
            manageAction +
          "</div>" +
        "</div>" +
      "</article>";
    }).join("");
  }

  function renderPolls() {
    var list = q("poll-list");
    var form = q("poll-form");
    var room = activeRoom();
    var items = (state.polls || []).slice(0);
    q("polls-status-pill").textContent = items.length + " enquetes";
    form.classList.toggle("hidden", !room || isHubView() || accessForRoom(room) === "locked" || accessForRoom(room) === "vip" || (state.viewer && state.viewer.role === "ai"));
    if (!room || isHubView()) {
      list.innerHTML = "<div class='empty-state'>Escolhe uma sala de conversa para abrir ou votar em enquetes.</div>";
      return;
    }
    if (isAppsLabRoom(room)) {
      form.classList.add("hidden");
      list.innerHTML = "<div class='empty-state'>Apps Lab agora e central do app. Enquete ficou fora dessa area.</div>";
      return;
    }
    if (accessForRoom(room) === "locked" || accessForRoom(room) === "vip") {
      list.innerHTML = "<div class='empty-state'>Quando a sala abrir, as enquetes aparecem aqui.</div>";
      return;
    }
    if (!items.length) {
      list.innerHTML = "<div class='empty-state'>Ainda nao tem enquete nessa sala. Cria uma e decide o rumo da bagunca.</div>";
      return;
    }
    list.innerHTML = items.map(function(poll) {
      var manageAction = canManageOwnedContent(poll.createdBy)
        ? "<button class='btn btn-ghost btn-danger' type='button' data-action='delete-poll' data-poll-id='" + Number(poll.id) + "'>Apagar</button>"
        : "";
      return "<article class='poll-card'>" +
        "<div class='utility-head'>" +
          "<h4>" + esc(poll.question) + "</h4>" +
          "<div class='inline-row'>" +
            "<span class='ghost-pill'>" + esc((poll.totalVotes || 0) + " votos") + "</span>" +
            manageAction +
          "</div>" +
        "</div>" +
        "<p class='media-meta'>" + esc((poll.createdByName || "alguem") + " // " + formatDateTime(poll.createdAt)) + "</p>" +
        "<div class='poll-options'>" +
          (poll.options || []).map(function(option) {
            var percent = poll.totalVotes ? Math.round((Number(option.votes || 0) / Number(poll.totalVotes || 1)) * 100) : 0;
            return "<button type='button' class='poll-option" + (option.viewerVoted ? " active" : "") + "' data-action='vote-poll' data-poll-id='" + Number(poll.id) + "' data-option-id='" + Number(option.id) + "'>" +
              "<div class='poll-option-copy'>" +
                "<strong>" + esc(option.label) + "</strong>" +
                "<span>" + esc((option.votes || 0) + " voto(s) // " + percent + "%") + "</span>" +
              "</div>" +
              "<span class='poll-option-check'>" + (option.viewerVoted ? "votado" : "votar") + "</span>" +
            "</button>";
          }).join("") +
        "</div>" +
      "</article>";
    }).join("");
  }

  function momentCandidates() {
    var base = isHubView()
      ? Object.keys(state.latestByRoom || {}).map(function(roomId) { return state.latestByRoom[roomId]; }).filter(Boolean)
      : currentRoomMessages().slice(0);
    return base.filter(function(message) {
      return message && !message.blockedByViewer;
    }).sort(function(a, b) {
      function score(message) {
        var reactionScore = (message.reactions || []).reduce(function(sum, item) { return sum + Number(item.count || 0); }, 0);
        return (message.isPinned ? 8 : 0) + (message.viewerFavorited ? 4 : 0) + reactionScore + (message.attachment ? 2 : 0);
      }
      return score(b) - score(a) || (new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    }).slice(0, 4);
  }

  function renderMoments() {
    var list = q("moments-list");
    var items = momentCandidates();
    q("moments-status-pill").textContent = items.length + " destaques";
    if (!items.length) {
      list.innerHTML = "<div class='empty-state'>Quando surgirem mensagens quentes, fixadas ou muito reagidas, elas aparecem aqui.</div>";
      return;
    }
    list.innerHTML = items.map(function(message) {
      var room = roomById(message.roomId);
      var preview = message.body || (message.attachment ? "[anexo] " + message.attachment.name : "[sem texto]");
      var reactionScore = (message.reactions || []).reduce(function(sum, item) { return sum + Number(item.count || 0); }, 0);
      return "<article class='favorite-card moment-card'>" +
        "<strong>" + esc((room ? displayRoomName(room) : "Sala") + " // " + (message.authorName || "Mensagem")) + "</strong>" +
        "<p>" + esc(preview.slice(0, 170)) + "</p>" +
        "<div class='inline-row'>" +
          (message.isPinned ? "<span class='ghost-pill'>fixada</span>" : "") +
          (message.attachment ? "<span class='ghost-pill'>" + esc(message.attachment.kind || "midia") + "</span>" : "") +
          "<span class='ghost-pill'>" + esc(reactionScore + " reacoes") + "</span>" +
        "</div>" +
        "<button class='btn btn-ghost' type='button' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Abrir destaque</button>" +
      "</article>";
    }).join("");
  }

  function renderUnreadBanner() {
    var wrap = q("unread-banner");
    var marker = state.latestUnreadByRoom[String(state.activeRoomId)];
    if (isHubView() || isAppsLabRoom(activeRoom()) || !marker) {
      wrap.classList.add("hidden");
      wrap.innerHTML = "";
      return;
    }
    wrap.classList.remove("hidden");
    wrap.innerHTML =
      "<span>Tem novidade nessa conversa desde tua ultima passada.</span>" +
      "<button type='button' class='btn btn-ghost' data-action='jump-unread' data-room-id='" + Number(state.activeRoomId) + "' data-message-id='" + Number(marker) + "'>Ir pras novidades</button>";
  }

  function renderFiles() {
    var list = q("room-media-list");
    var attachments = filteredRoomAttachments();
    var total = currentRoomAttachments().length;
    var summary = q("media-summary");
    var room = activeRoom();
    var totalBytes = totalAttachmentBytes(currentRoomAttachments());
    renderRoomSpotlight();
    q("media-search-input").value = state.mediaSearch;
    q("media-sort-select").value = state.mediaSort;
    Array.prototype.slice.call(document.querySelectorAll("[data-media-filter]")).forEach(function(item) {
      item.classList.toggle("active", item.getAttribute("data-media-filter") === state.mediaFilter);
    });
    if (isHubView()) {
      q("files-status-pill").textContent = "0 itens";
      renderMediaInsights(null, [], 0);
      if (summary) {
        summary.textContent = "Seleciona uma sala real para abrir a biblioteca de midia.";
      }
      list.innerHTML = renderContextEmptyCard(
        "Biblioteca sem sala",
        "Seleciona uma sala de chat para ver a midia dela e abrir a biblioteca do recorte certo.",
        "<span class='ghost-pill'>midia</span><span class='ghost-pill'>por sala</span>",
        "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='members'>Membros</button>" +
        "<button class='btn btn-ghost' type='button' data-action='open-inspector-section' data-section='search'>Busca</button>"
      );
      return;
    }
    if (isAppsLabRoom(activeRoom())) {
      q("files-status-pill").textContent = "app only";
      renderMediaInsights(null, [], 0);
      if (summary) {
        summary.textContent = "Apps Lab agora virou central do UniversalD. Sem biblioteca de midia misturada aqui.";
      }
      list.innerHTML = renderContextEmptyCard(
        "Apps Lab dedicado",
        "Usa essa area para abrir, baixar e atualizar o app oficial, sem biblioteca de midia misturada aqui.",
        "<span class='ghost-pill'>app only</span><span class='ghost-pill'>central</span>",
        "<button class='btn btn-ghost' type='button' data-action='open-app-center'>Central do App</button>"
      );
      return;
    }
    q("files-status-pill").textContent = attachments.length + " / " + total + " itens";
    renderMediaInsights(room, attachments, total);
    list.innerHTML = "";
    list.className = "media-list";
    if (room && room.slug === "fotos") {
      list.classList.add("gallery-mode");
    } else {
      list.classList.add("file-mode");
    }
    if (summary) {
      summary.textContent = (room && room.slug === "fotos" ? "Galeria viva" : "Biblioteca da sala") + " // filtro " + (state.mediaFilter === "all" ? "geral" : state.mediaFilter) + " // ordem " + mediaSortLabel(state.mediaSort) + " // " + attachments.length + " visiveis // peso " + bytesLabel(totalBytes);
    }
    if (!attachments.length) {
      list.innerHTML = renderContextEmptyCard(
        "Nada bateu no filtro",
        "Nenhum anexo bate com esse filtro nessa sala. Tenta limpar a busca, trocar a ordem ou mandar novo conteudo.",
        "<span class='ghost-pill'>" + esc(state.mediaFilter === "all" ? "geral" : state.mediaFilter) + "</span><span class='ghost-pill'>" + esc(mediaSortLabel(state.mediaSort)) + "</span>",
        "<button class='btn btn-ghost' type='button' data-action='trigger-attach'>Mandar arquivo</button>" +
        "<button class='btn btn-ghost' type='button' data-action='focus-composer'>Abrir composer</button>"
      );
      return;
    }
    attachments.forEach(function(message) {
      var attachment = message.attachment;
      var attachmentUrl = safeAttachmentUrl(attachment.url);
      var card = document.createElement("article");
      card.className = "media-card";
      var size = bytesLabel(attachment.sizeBytes);
      var dims = attachment.width && attachment.height ? (attachment.width + "x" + attachment.height) : "";
      var kindCopy = String(attachment.kind || "arquivo");
      var excerpt = String(message.body || "").trim();
      card.innerHTML =
        "<div class='media-card-head'><strong>" + esc(attachment.name) + "</strong><span class='ghost-pill kind-pill'>" + esc(attachment.kind || "arquivo") + "</span></div>" +
        (attachmentUrl && attachment.kind === "image" ? "<button type='button' class='media-thumb' data-action='preview-attachment' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'><img alt='" + esc(attachment.name) + "' src='" + esc(attachmentUrl) + "' /></button>" : "") +
        (attachmentUrl && attachment.kind === "video" ? "<button type='button' class='media-thumb' data-action='preview-attachment' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'><video preload='metadata' muted src='" + esc(attachmentUrl) + "'></video></button>" : "") +
        (attachmentUrl && attachment.kind === "audio" ? "<audio controls preload='metadata' src='" + esc(attachmentUrl) + "'></audio>" : "") +
        (attachment.kind === "file" ? "<div class='media-file-fallback'><strong>" + esc(attachment.extension || "arquivo") + "</strong><span>" + esc(attachment.contentType || "download") + "</span></div>" : "") +
        "<div class='inline-row media-card-pills'>" +
          "<span class='ghost-pill'>" + esc(size) + "</span>" +
          (dims ? "<span class='ghost-pill'>" + esc(dims) + "</span>" : "") +
          "<span class='ghost-pill'>" + esc(message.authorName || "alguem") + "</span>" +
          "<span class='ghost-pill'>" + esc(formatRelative(message.createdAt)) + "</span>" +
        "</div>" +
        (excerpt ? "<p class='media-caption'>" + esc(excerpt.slice(0, 140)) + "</p>" : "") +
        "<div class='media-card-origin'>" +
          "<div class='media-card-origin-copy'>" +
            "<strong>" + esc(room ? displayRoomName(room) : "Sem sala") + "</strong>" +
            "<span>" + esc(kindCopy + " // " + bytesLabel(attachment.sizeBytes)) + "</span>" +
          "</div>" +
          "<div class='inline-row media-card-flow'>" +
            (attachment.extension ? "<span class='ghost-pill'>" + esc(String(attachment.extension).toUpperCase()) + "</span>" : "") +
            (attachment.contentType ? "<span class='ghost-pill'>" + esc(attachment.contentType) + "</span>" : "") +
          "</div>" +
        "</div>" +
        "<div class='attachment-actions'>" +
          "<button class='btn btn-primary' type='button' data-action='preview-attachment' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Abrir</button>" +
          "<button class='btn btn-ghost' type='button' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Origem</button>" +
          (attachmentUrl ? "<a class='btn btn-ghost' href='" + esc(attachmentUrl) + "' target='_blank' rel='noreferrer'>Baixar</a>" : "") +
        "</div>";
      list.appendChild(card);
    });
  }

  function renderLogs() {
    var list = q("logs-list");
    var canSee = state.viewer && (state.viewer.role === "owner" || state.viewer.role === "admin");
    list.innerHTML = "";
    if (!canSee) {
      list.innerHTML = "<div class='empty-state'>Logs liberados so pra admin e owner.</div>";
      return;
    }
    if (!state.recentLogs.length) {
      list.innerHTML = "<div class='empty-state'>Sem logs carregados.</div>";
      return;
    }
    state.recentLogs.forEach(function(item) {
      var card = document.createElement("article");
      card.className = "log-item";
      card.innerHTML =
        "<strong>" + esc(item.actorName || "sistema") + " // " + esc(item.action || "acao") + "</strong>" +
        "<p>" + esc(item.detail || "sem detalhe") + "</p>" +
        "<span class='message-time'>" + esc(formatDateTime(item.createdAt)) + "</span>";
      list.appendChild(card);
    });
  }

  function renderUserRoster() {
    var list = q("user-roster");
    q("user-roster-status-pill").textContent = state.users.length + " usuarios";
    list.innerHTML = "";
    if (!state.users.length) {
      list.innerHTML = "<div class='empty-state'>Sem usuarios extras cadastrados ainda.</div>";
      return;
    }
    state.users.forEach(function(user) {
      var manageable = state.viewer && state.viewer.role === "owner" && Number(user.id) !== Number(state.viewer.id) && user.role !== "owner" && user.role !== "ai";
      var card = document.createElement("article");
      card.className = "roster-item";
      card.innerHTML =
        "<div class='roster-copy'>" +
          "<strong>" + esc(user.displayName || user.username) + "</strong>" +
          "<span>" + esc((user.role || "member") + " // " + (user.email || "sem email")) + "</span>" +
          "<span>Ultimo login: " + esc(user.lastLoginAt ? formatDateTime(user.lastLoginAt) : "nunca") + "</span>" +
        "</div>" +
        (manageable
          ? "<div class='roster-actions'>" +
              "<select data-role-select='" + Number(user.id) + "'>" +
                "<option value='member'" + (user.role === "member" ? " selected" : "") + ">membro</option>" +
                "<option value='vip'" + (user.role === "vip" ? " selected" : "") + ">vip</option>" +
                "<option value='admin'" + (user.role === "admin" ? " selected" : "") + ">admin</option>" +
              "</select>" +
              "<button class='btn btn-ghost' type='button' data-action='apply-user-role' data-user-id='" + Number(user.id) + "'>Salvar cargo</button>" +
              "<button class='btn btn-danger' type='button' data-action='expel-user' data-user-id='" + Number(user.id) + "'>Expulsar</button>" +
            "</div>"
          : "<div class='roster-actions'><span class='ghost-pill'>" + esc(user.role || "member") + "</span></div>");
      list.appendChild(card);
    });
  }

  function renderJoinRequests() {
    var list = q("join-request-list");
    var pending = state.joinRequests.filter(function(item) {
      return String(item.status || "") === "pending";
    }).length;
    q("join-request-status-pill").textContent = pending + " pendentes";
    list.innerHTML = "";
    if (!state.joinRequests.length) {
      list.innerHTML = "<div class='empty-state'>Nenhum pedido de entrada apareceu ainda.</div>";
      return;
    }
    state.joinRequests.forEach(function(item) {
      var status = String(item.status || "pending");
      var card = document.createElement("article");
      card.className = "roster-item join-request-item status-" + status;
      card.innerHTML =
        "<div class='roster-copy'>" +
          "<strong>" + esc(item.displayName || item.email || "Pedido novo") + "</strong>" +
          "<span>" + esc(item.email || "sem email") + " // " + esc(status) + "</span>" +
          "<span>Pedido: " + esc(formatDateTime(item.requestedAt)) + "</span>" +
          (item.note ? "<p class='roster-note'>" + esc(item.note) + "</p>" : "") +
          (item.reviewNote ? "<p class='roster-note'><strong>Retorno:</strong> " + esc(item.reviewNote) + "</p>" : "") +
          (item.accessCode ? "<p class='roster-note'><strong>Codigo:</strong> <code>" + esc(item.accessCode) + "</code>" + (item.emailSent ? " // enviado por email" : " // enviar manualmente") + "</p>" : "") +
        "</div>" +
        "<div class='roster-actions'>" +
          (status === "pending"
            ? "<button class='btn btn-primary' type='button' data-action='review-join-request' data-request-id='" + Number(item.id) + "' data-approve='true'>Aprovar</button>" +
              "<button class='btn btn-danger' type='button' data-action='review-join-request' data-request-id='" + Number(item.id) + "' data-approve='false'>Recusar</button>"
            : "<span class='ghost-pill'>" + esc(status) + "</span>") +
        "</div>";
      list.appendChild(card);
    });
    renderDashboardPulse();
  }

  function renderSearchResults() {
    var wrap = q("search-results");
    var summary = q("search-summary");
    var roomCount;
    var userCount;
    var messageCount;
    wrap.innerHTML = "";
    if (!state.searchQuery) {
      summary.classList.add("hidden");
      summary.textContent = "";
      wrap.innerHTML = "<div class='empty-state'>Digite algo e roda a busca global.</div>";
      return;
    }
    if (!state.searchResult) {
      summary.classList.remove("hidden");
      summary.textContent = "Rodando busca em salas, usuarios e mensagens...";
      wrap.innerHTML = "<div class='empty-state'>Buscando...</div>";
      return;
    }
    roomCount = (state.searchResult.rooms || []).length;
    userCount = (state.searchResult.users || []).length;
    messageCount = (state.searchResult.messages || []).length;
    summary.classList.remove("hidden");
    summary.textContent = "\"" + state.searchQuery + "\" // " + (roomCount + userCount + messageCount) + " resultados // salas " + roomCount + " // usuarios " + userCount + " // mensagens " + messageCount;
    renderSearchSection(wrap, "Salas", state.searchResult.rooms || [], function(room) {
      return "<button class='btn btn-ghost' type='button' data-action='jump-room' data-room-id='" + Number(room.id) + "'>Abrir sala</button>";
    }, function(room) {
      return "<div class='search-card-head'><strong>" + esc(displayRoomName(room)) + "</strong><span class='ghost-pill'>" + esc(roomKindLabel(room)) + "</span></div><p>" + esc(displayRoomDescription(room)) + "</p>";
    });
    renderSearchSection(wrap, "Usuarios", state.searchResult.users || [], function(user) {
      var actions = "<button class='btn btn-ghost' type='button' data-action='open-user-profile' data-user-id='" + Number(user.id) + "'>Perfil</button>";
      if (!state.viewer || Number(user.id) === Number(state.viewer.id) || user.role === "ai") {
        return actions;
      }
      return actions + "<button class='btn btn-ghost' type='button' data-action='open-dm' data-user-id='" + Number(user.id) + "'>Abrir DM</button>";
    }, function(user) {
      var presence = presenceByUserId(user.id) || {};
      return "<div class='search-card-head'><strong>" + esc(user.displayName || user.username) + "</strong><span class='ghost-pill'>" + esc(user.role || "member") + "</span></div><p>" + esc((presence.online ? (presence.status || "online") : "offline") + " // " + (user.email || "sem email")) + "</p>";
    });
    renderSearchSection(wrap, "Mensagens", state.searchResult.messages || [], function(message) {
      return "<button class='btn btn-ghost' type='button' data-action='jump-message' data-room-id='" + Number(message.roomId) + "' data-message-id='" + Number(message.id) + "'>Ir pra mensagem</button>";
    }, function(message) {
      var preview = message.body || (message.attachment ? "[anexo] " + message.attachment.name : "[sem texto]");
      var meta = message.attachment ? (" // " + (message.attachment.kind || "arquivo")) : "";
      var room = roomById(message.roomId);
      return "<div class='search-card-head'><strong>" + esc(message.authorName || "Mensagem") + "</strong><span class='ghost-pill'>" + esc(room ? displayRoomName(room) : "sala") + "</span></div><p>" + esc(preview + meta) + "</p><span class='search-card-meta'>" + esc(formatRelative(message.createdAt)) + "</span>";
    });
  }

  function renderSearchSection(root, title, items, actionMarkup, copyMarkup) {
    var section = document.createElement("section");
    section.className = "search-section";
    section.innerHTML = "<div class='search-section-head'><h4>" + esc(title) + "</h4><span class='ghost-pill'>" + items.length + "</span></div>";
    if (!items.length) {
      section.innerHTML += "<div class='empty-state'>Sem resultado nessa faixa.</div>";
      root.appendChild(section);
      return;
    }
    items.forEach(function(item) {
      var card = document.createElement("article");
      card.className = "search-card";
      card.innerHTML = copyMarkup(item) + actionMarkup(item);
      section.appendChild(card);
    });
    root.appendChild(section);
  }

  function renderAdminPanels() {
    var privileged = state.viewer && (state.viewer.role === "owner" || state.viewer.role === "admin");
    var isOwner = state.viewer && state.viewer.role === "owner";
    var room = activeRoom();
    q("inspector-section-logs").classList.toggle("hidden", !privileged);
    q("inspector-section-admin").classList.toggle("hidden", !privileged);
    q("terminal-card").classList.toggle("hidden", !(privileged && room && room.slug === "apps-lab"));
    q("owner-card").classList.toggle("hidden", !isOwner);
    if (!privileged && (state.inspectorTab === "logs" || state.inspectorTab === "admin")) {
      state.inspectorTab = "overview";
    }
  }

  function renderInspectorTabs() {
    var map = {
      overview: "inspector-section-overview",
      members: "inspector-section-members",
      files: "inspector-section-files",
      search: "inspector-section-search",
      logs: "inspector-section-logs",
      admin: "inspector-section-admin"
    };
    document.querySelectorAll(".inspector-section").forEach(function(panel) {
      var active = panel.id === map[state.inspectorTab];
      panel.classList.toggle("focused", active);
      panel.classList.toggle("hidden-panel", !active);
    });
    renderSearchResults();
  }

  async function loadMessages(roomId, silent) {
    var room = roomById(roomId);
    if (!room || accessForRoom(room) === "locked" || accessForRoom(room) === "vip") {
      renderMessages(roomId);
      renderUnreadBanner();
      renderFiles();
      state.polls = [];
      renderPolls();
      return;
    }
    try {
      var data = await apiFetch("/api/panel/messages?roomId=" + encodeURIComponent(roomId) + "&limit=80");
      state.messagesByRoom[String(roomId)] = data.messages || [];
      state.pinnedByRoom[String(roomId)] = data.pins || [];
      state.unread[String(roomId)] = 0;
      renderRooms();
      renderMessages(roomId);
      renderUnreadBanner();
      renderMoments();
      renderFiles();
      renderPinnedStrip();
      renderFavorites();
      if (!silent) {
        pushActivity("Sala " + room.name + " carregada.", "ok");
      }
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function loadPolls(roomId, silent) {
    var room = roomById(roomId);
    if (!room || accessForRoom(room) === "locked" || accessForRoom(room) === "vip") {
      state.polls = [];
      renderPolls();
      return;
    }
    try {
      var data = await apiFetch("/api/panel/polls?roomId=" + encodeURIComponent(roomId) + "&limit=12");
      state.polls = data.polls || [];
      renderPolls();
      if (!silent) {
        pushActivity("Enquetes da sala " + displayRoomName(room) + " carregadas.", "ok");
      }
    } catch (err) {
      state.polls = [];
      renderPolls();
      if (!silent) {
        toast(err.message, "err");
      }
    }
  }

  async function selectRoom(roomId, silent, highlightMessageId) {
    stopTyping(false);
    if (Number(state.activeRoomId) !== Number(roomId)) {
      state.editingMessage = null;
      nudgeConversation();
    }
    state.activeRoomId = Number(roomId);
    state.activeNavId = navIdForRoom(activeRoom());
    state.highlightMessageId = Number(highlightMessageId || 0);
    saveStoredActiveRoomId(state.activeRoomId);
    saveStoredActiveNavId(state.activeNavId);
    renderRooms();
    renderRails();
    ensureActiveNavVisible(true);
    renderHeaderAndRoomState();
    renderDashboard();
    renderAppsLabDeck();
    renderActivity();
    renderEvents();
    renderMoments();
    renderUnreadBanner();
    renderPendingAttachment();
    renderReplyChip();
    restoreComposerDraft();
    if (state.isMobile) {
      state.utilityStripCollapsed = true;
      state.chatContextCollapsed = true;
      syncUtilityStripUI();
      applyChatContextState();
    }
    renderAdminPanels();
    await loadMessages(roomId, silent);
    await loadPolls(roomId, true);
    sendPresence();
    if (state.compactLayout) {
      closeSidebar();
      closeInspector();
    }
  }

  async function selectPrimaryNav(navId, silent) {
    var nav = navDefinition(navId);
    var room = navRoom(nav);
    stopTyping(false);
    state.editingMessage = null;
    state.activeNavId = nav.id;
    state.highlightMessageId = 0;
    saveStoredActiveNavId(state.activeNavId);
    nudgeConversation();
    if (nav.kind === "sus") {
      state.activeRoomId = 0;
      saveStoredActiveRoomId(0);
      renderRooms();
      renderRails();
      ensureActiveNavVisible(true);
      renderHeaderAndRoomState();
      renderDashboard();
      renderAppsLabDeck();
      renderSusPanel();
      renderActivity();
      renderEvents();
      renderMoments();
      renderUnreadBanner();
      renderPendingAttachment();
      renderReplyChip();
      renderEditChip();
      renderMessages(0);
      renderFiles();
      state.polls = [];
      renderPolls();
      renderAdminPanels();
      renderInspectorTabs();
      if (state.isMobile) {
        state.utilityStripCollapsed = true;
        state.chatContextCollapsed = true;
        syncUtilityStripUI();
        applyChatContextState();
      }
      sendPresence();
      if (!silent) {
        toast("Bah... tu abriu a aba suspeita. Depois nao reclama.", "warn");
      }
      if (state.compactLayout) {
        closeSidebar();
        closeInspector();
      }
      return;
    }
    if (nav.kind === "dm" && !room) {
      state.activeRoomId = 0;
      saveStoredActiveRoomId(0);
      renderRooms();
      renderRails();
      ensureActiveNavVisible(true);
      renderHeaderAndRoomState();
      renderDashboard();
      renderAppsLabDeck();
      renderActivity();
      renderEvents();
      renderMoments();
      renderUnreadBanner();
      renderPendingAttachment();
      renderReplyChip();
      restoreComposerDraft();
      renderMessages(0);
      renderFiles();
      state.polls = [];
      renderPolls();
      renderAdminPanels();
      renderInspectorTabs();
      if (state.isMobile) {
        state.utilityStripCollapsed = true;
        state.chatContextCollapsed = true;
        syncUtilityStripUI();
        applyChatContextState();
      }
      sendPresence();
      if (state.compactLayout) {
        closeSidebar();
        closeInspector();
      }
      return;
    }
    if (room) {
      await selectRoom(room.id, silent);
      return;
    }
    toast("Essa area ainda nao tem sala ligada.", "warn");
  }

  function markRoomActivity(roomId, message) {
    var room = roomById(roomId);
    if (!room || !message) {
      return;
    }
    room.lastMessageAt = message.createdAt;
    room.lastMessagePreview = message.body || (message.attachment ? message.attachment.name : "");
  }

  async function handleLogin(event) {
    event.preventDefault();
    var loginStage = q("login-view").querySelector(".login-stage");
    var submit = q("login-submit");
    var login = q("login-input").value.trim();
    var password = q("password-input").value.trim();
    if (!login || !password) {
      q("login-error").classList.remove("hidden");
      q("login-error").textContent = "Bah, preenche usuario e senha antes de posar de hacker.";
      pulseClass(loginStage, "login-error-pulse");
      playTone("err");
      return;
    }
    try {
      setButtonBusy(submit, true, "Entrando...", "Entrar no Painel");
      var data = await apiFetch("/api/panel/login", {
        method: "POST",
        body: JSON.stringify({ login: login, password: password })
      });
      storeSessionId(data && data.sessionId);
      q("login-input").value = "";
      q("password-input").value = "";
      q("login-error").classList.add("hidden");
      pulseClass(loginStage, "login-success-pulse");
      await wait(320);
      applyBootstrap(data.bootstrap);
      showPanel();
      toast(pickRandom(LOGIN_SUCCESS_LINES) || "Acesso liberado ao Painel Dief.", "ok");
    } catch (err) {
      var errorText = pickRandom(LOGIN_ERROR_LINES) || "Acesso negado.";
      var rawMessage = String(err.message || "");
      var isCooldown = /segura .*tentar logar/i.test(rawMessage);
      q("login-error").classList.remove("hidden");
      q("login-error").textContent = isCooldown
        ? rawMessage
        : (/credenciais|senha|login/i.test(rawMessage) ? errorText : (errorText + " " + rawMessage));
      pulseClass(loginStage, "login-error-pulse");
      playTone("err");
    } finally {
      setButtonBusy(submit, false, "Entrando...", "Entrar no Painel");
    }
  }

  async function handleLogout() {
    var button = q("btn-logout");
    try {
      setButtonBusy(button, true, "Saindo...", "Sair");
      await apiFetch("/api/panel/logout", { method: "POST", body: JSON.stringify({}) });
    } catch (err) {}
    setButtonBusy(button, false, "Saindo...", "Sair");
    clearStoredSessionId();
    resetSession();
    showLogin("Tu saiu do painel.");
  }

  async function handleComposerSubmit(event) {
    event.preventDefault();
    var room = activeRoom();
    var body = q("composer-input").value.trim();
    var queuedAttachments = readyPendingUploads();
    if (state.sendingMessage) {
      return;
    }
    if (hasUploadingPendingUploads()) {
      toast("Espera os uploads terminarem antes de mandar a mensagem.", "warn");
      return;
    }
    if (hasErroredPendingUploads()) {
      if (!body && !queuedAttachments.length) {
        toast("Tem upload falhando. Remove ou tenta de novo antes de mandar algo vazio.", "warn");
        return;
      }
      clearErroredPendingUploads();
      queuedAttachments = readyPendingUploads();
      toast("Limpei os uploads que falharam e segui com o resto.", "warn");
    }
    if (isSusView()) {
      toast("Nessa aba tu nao conversa. Tu so provoca o destino.", "warn");
      return;
    }
    if (isAppsLabRoom(room)) {
      toast("Apps Lab virou central do app. Aqui tu abre, baixa e atualiza sem chat no meio.", "warn");
      return;
    }
    if (!room) {
      return;
    }
    if (body.charAt(0) === "/" && !queuedAttachments.length) {
      await handleSlashCommand(body);
      return;
    }
    if (accessForRoom(room) === "locked") {
      openUnlockModal(room.id);
      return;
    }
    if (accessForRoom(room) === "vip") {
      toast("Essa sala nao te quis hoje.", "warn");
      return;
    }
    if (!body && !queuedAttachments.length && !(state.editingMessage && (state.editingMessage.attachment || state.editingMessage.hasAttachment))) {
      toast("Mensagem vazia nao vale, tche.", "warn");
      return;
    }

    try {
      state.sendingMessage = true;
      renderHeaderAndRoomState();
      if (state.editingMessage) {
        if (queuedAttachments.length) {
          toast("Edicao nao aceita anexo novo junto. Fecha a edicao ou manda separado.", "warn");
          return;
        }
        var edited = await apiFetch("/api/panel/messages", {
          method: "PUT",
          body: JSON.stringify({
            roomId: room.id,
            messageId: state.editingMessage.id,
            body: body
          })
        });
        replaceRoomMessage(room.id, edited.message);
        q("composer-input").value = "";
        clearRoomDraft(room.id);
        state.editingMessage = null;
        renderEditChip();
        renderHeaderAndRoomState();
        renderRooms();
        renderMessages(room.id);
        renderPinnedStrip();
        renderFavorites();
        renderMoments();
        toast("Mensagem editada sem quebrar o fio.", "ok");
        return;
      }
      if (room.slug === "nego-dramias-ia" && !queuedAttachments.length) {
        var aiData = await apiFetch("/api/panel/ai/chat", {
          method: "POST",
          body: JSON.stringify({ roomId: room.id, prompt: body })
        });
        state.messagesByRoom[String(room.id)] = (state.messagesByRoom[String(room.id)] || []).concat([aiData.question, aiData.reply]);
        state.latestByRoom[String(room.id)] = aiData.reply;
        markRoomActivity(room.id, aiData.reply);
        delete state.latestUnreadByRoom[String(room.id)];
        q("composer-input").value = "";
        clearRoomDraft(room.id);
        state.replyTarget = null;
        state.editingMessage = null;
        stopTyping(false);
        renderReplyChip();
        renderEditChip();
        renderRooms();
        renderMessages(room.id);
        renderUnreadBanner();
        renderDashboard();
        renderFiles();
        renderMoments();
        toast("Nego Dramias respondeu no sotaque.", "ok");
        return;
      }

      var sentMessages = [];
      if (queuedAttachments.length) {
        for (var i = 0; i < queuedAttachments.length; i++) {
          var queued = queuedAttachments[i].attachment;
          var resultWithAttachment = await apiFetch("/api/panel/messages", {
            method: "POST",
            body: JSON.stringify({
              roomId: room.id,
              body: i === 0 ? body : "",
              kind: queued.kind || "file",
              attachment: queued,
              replyToId: i === 0 && state.replyTarget ? state.replyTarget.id : 0
            })
          });
          sentMessages.push(resultWithAttachment.message);
        }
      } else {
        var result = await apiFetch("/api/panel/messages", {
          method: "POST",
          body: JSON.stringify({
            roomId: room.id,
            body: body,
            kind: "text",
            attachment: null,
            replyToId: state.replyTarget ? state.replyTarget.id : 0
          })
        });
        sentMessages.push(result.message);
      }
      if (!state.messagesByRoom[String(room.id)]) {
        state.messagesByRoom[String(room.id)] = [];
      }
      sentMessages.forEach(function(msg) {
        state.messagesByRoom[String(room.id)].push(msg);
        state.latestByRoom[String(room.id)] = msg;
        markRoomActivity(room.id, msg);
      });
      state.pendingUploads = [];
      state.pendingAttachment = null;
      state.replyTarget = null;
      state.editingMessage = null;
      delete state.latestUnreadByRoom[String(room.id)];
      q("composer-input").value = "";
      clearRoomDraft(room.id);
      stopTyping(false);
      renderPendingAttachment();
      renderReplyChip();
      renderEditChip();
      renderRooms();
      renderMessages(room.id);
      renderUnreadBanner();
      renderDashboard();
      renderFiles();
      renderFavorites();
      renderMoments();
      toast(queuedAttachments.length ? "Midia enviada e salva no historico." : "Mensagem enviada sem peidar no motor.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      state.sendingMessage = false;
      renderHeaderAndRoomState();
    }
  }

  async function handleSlashCommand(text) {
    var input = text.trim();
    var room = activeRoom();
    if (input === "/help") {
      toast("Comandos: /help /online /matrix /obsidian /ember /cobalt /neon /logs /search termo /ping /clear", "ok");
      return;
    }
    if (input === "/online") {
      toast("Tem " + state.online.filter(function(item) { return item.online; }).length + " gente online agora.", "ok");
      return;
    }
    if (input === "/clear") {
      q("composer-input").value = "";
      if (room) {
        clearRoomDraft(room.id);
      }
      state.replyTarget = null;
      state.pendingUploads = [];
      state.pendingAttachment = null;
      renderPendingAttachment();
      renderReplyChip();
      syncComposerDraftHint();
      stopTyping(true);
      return;
    }
    if (input === "/logs") {
      setInspectorTab("logs");
      if (state.viewer && (state.viewer.role === "owner" || state.viewer.role === "admin")) {
        loadLogs();
      }
      return;
    }
    if (input.indexOf("/search ") === 0) {
      var query = input.slice(8).trim();
      if (!query) {
        toast("Manda um termo depois do /search.", "warn");
        return;
      }
      await runSearch(query);
      return;
    }
    if (input === "/ping") {
      var start = Date.now();
      await apiFetch("/api/panel/bootstrap");
      toast("Ping do painel: " + (Date.now() - start) + "ms.", "ok");
      return;
    }
    if (["/matrix", "/obsidian", "/ember", "/cobalt", "/neon"].indexOf(input) >= 0) {
      await saveProfile({
        displayName: state.viewer.displayName,
        bio: state.viewer.bio,
        theme: input.slice(1),
        accentColor: state.viewer.accentColor,
        avatarUrl: state.viewer.avatarUrl,
        status: state.viewer.status,
        statusText: state.viewer.statusText || ""
      }, "Tema trocado pra " + input.slice(1) + ".");
      return;
    }
    if (input === "/ai" && room && room.slug !== "nego-dramias-ia") {
      var aiRoom = state.rooms.find(function(item) { return item.slug === "nego-dramias-ia"; });
      if (aiRoom) {
        await selectRoom(aiRoom.id);
      }
      return;
    }
    toast("Comando desconhecido. Ate o Nego Dramias te julgou.", "warn");
  }

  function syncBootstrapSnapshot(refreshed) {
    state.viewer = refreshed.viewer || state.viewer;
    state.rooms = refreshed.rooms || state.rooms;
    state.roomAccess = refreshed.roomAccess || state.roomAccess;
    state.online = refreshed.online || state.online;
    state.typing = refreshed.typing || state.typing;
    state.recentLogs = refreshed.recentLogs || state.recentLogs;
    state.latestByRoom = latestMessagesMap(refreshed.latestMessages || []);
    state.events = refreshed.events || state.events;
    state.roomVersions = refreshed.roomVersions || state.roomVersions;
    state.blockedUserIds = (refreshed.blockedUserIds || state.blockedUserIds || []).map(Number);
    state.mutedUserIds = (refreshed.mutedUserIds || state.mutedUserIds || []).map(Number);
    state.version = Number(refreshed.version || state.version || 0);
    state.previousOnlineMap = onlineMap(state.online);
    applyTheme();
    renderShell();
    updateDocumentTitle();
    if (state.viewer && state.viewer.role === "owner") {
      loadUsers();
      loadJoinRequests();
    }
  }

  async function uploadSelectedFile(input, onDone) {
    var uploaded;
    if (!input.files || !input.files.length) {
      return;
    }
    try {
      toast("Subindo avatar pro painel.", "warn");
      uploaded = await queueFilesForUpload([input.files[0]], { mode: "avatar" });
      if (uploaded[0] && onDone) {
        onDone(uploaded[0]);
      }
      toast("Upload pronto. Agora pode salvar o perfil.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      input.value = "";
    }
  }

  function openUnlockModal(roomId) {
    state.unlockRoomId = Number(roomId);
    var room = roomById(roomId);
    q("unlock-room-name").textContent = "Digite a senha da sala " + (room ? room.name : "");
    q("unlock-modal").classList.remove("hidden");
    syncTransientLayoutState();
    q("unlock-password").focus();
  }

  async function handleUnlockSubmit(event) {
    event.preventDefault();
    try {
      await apiFetch("/api/panel/rooms/unlock", {
        method: "POST",
        body: JSON.stringify({ roomId: state.unlockRoomId, password: q("unlock-password").value })
      });
      q("unlock-password").value = "";
      closeModal("unlock-modal");
      var refreshed = await apiFetch("/api/panel/bootstrap");
      syncBootstrapSnapshot(refreshed);
      await loadMessages(state.unlockRoomId, true);
      toast("Sala aberta. Agora entra sem fazer fiasco.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  function applyThemePreset(theme, accent) {
    q("profile-theme").value = theme || "matrix";
    if (accent) {
      q("profile-accent").value = accent;
    }
    syncThemePresetState();
  }

  function applyBannerPreset(bannerPreset) {
    q("profile-banner-preset").value = bannerPreset || "grid";
    syncBannerPresetState();
  }

  function openProfileModal() {
    if (!state.viewer) {
      return;
    }
    if (state.isMobile) {
      closeSidebar();
      closeInspector();
      state.utilityStripCollapsed = true;
      syncUtilityStripUI();
    }
    q("profile-display").value = state.viewer.displayName || "";
    q("profile-status").value = state.viewer.status || "online";
    q("profile-theme").value = state.viewer.theme || "matrix";
    q("profile-banner-preset").value = state.viewer.bannerPreset || "grid";
    q("profile-banner-url").value = state.viewer.bannerUrl || "";
    q("profile-accent").value = state.viewer.accentColor || "#7bff00";
    q("profile-avatar").value = state.viewer.avatarUrl || "";
    q("profile-bio").value = state.viewer.bio || "";
    q("profile-status-text").value = state.viewer.statusText || "";
    syncThemePresetState();
    syncBannerPresetState();
    renderProfileFormPreview();
    q("profile-modal").classList.remove("hidden");
    q("profile-modal").scrollTop = 0;
    syncTransientLayoutState();
  }

  function renderProfileFormPreview() {
    var avatarWrap = q("profile-preview-avatar");
    var banner = q("profile-preview-banner");
    var insightWrap = q("profile-preview-insights");
    var avatarUrl = safeAvatarUrl(q("profile-avatar").value.trim());
    var bannerUrl = safeBannerUrl(q("profile-banner-url").value.trim());
    var displayName = q("profile-display").value.trim() || (state.viewer && (state.viewer.displayName || state.viewer.username)) || "Painel Dief";
    var status = q("profile-status").value || "online";
    var statusText = q("profile-status-text").value.trim();
    var bio = q("profile-bio").value.trim();
    var theme = q("profile-theme").value || "matrix";
    var bannerPreset = q("profile-banner-preset").value || (state.viewer && state.viewer.bannerPreset) || "grid";
    var accent = q("profile-accent").value || (state.viewer && state.viewer.accentColor) || themeAccent(theme);
    var badgesWrap = q("profile-preview-badges");
    if (!avatarWrap) {
      return;
    }
    if (banner) {
      banner.style.background = panelBannerStyle(theme, accent, bannerPreset, bannerUrl);
    }
    if (avatarUrl) {
      avatarWrap.innerHTML = "<img class='message-avatar profile-avatar-lg' alt='" + esc(displayName) + "' src='" + esc(avatarUrl) + "' />";
    } else {
      avatarWrap.textContent = initials(displayName);
    }
    q("profile-preview-name").textContent = displayName;
    q("profile-preview-meta").textContent = status + " // " + themeLabel(theme) + " // " + bannerLabel(bannerPreset);
    q("profile-preview-status").textContent = statusText || "Sem status customizado.";
    q("profile-preview-status").classList.toggle("hidden", !statusText);
    q("profile-preview-bio").textContent = bio || "Tua bio aparece aqui antes de salvar.";
    if (badgesWrap) {
      badgesWrap.innerHTML =
        "<span class='ghost-pill'>" + esc(themeLabel(theme)) + "</span>" +
        "<span class='ghost-pill'>" + esc(bannerUrl ? "banner proprio" : bannerLabel(bannerPreset)) + "</span>" +
        "<span class='ghost-pill'>" + esc(status || "online") + "</span>" +
        "<span class='ghost-pill'>" + esc(avatarUrl ? "avatar custom" : "avatar base") + "</span>";
    }
    if (insightWrap) {
      var stats = currentViewerLocalStats();
      insightWrap.innerHTML =
        "<article class='insight-card'><span>teu mapa</span><strong>" + stats.messageCount + "</strong><p>mensagens tuas visiveis agora</p></article>" +
        "<article class='insight-card'><span>midia</span><strong>" + stats.attachmentCount + "</strong><p>anexos teus no recorte local</p></article>" +
        "<article class='insight-card'><span>salas</span><strong>" + stats.roomCount + "</strong><p>salas onde tua marca apareceu</p></article>" +
        "<article class='insight-card'><span>tom</span><strong>" + esc(themeLabel(theme)) + "</strong><p>" + esc(statusText || "sem status customizado") + "</p></article>" +
        "<article class='insight-card'><span>capa</span><strong>" + esc(bannerUrl ? "Imagem propria" : bannerLabel(bannerPreset)) + "</strong><p>" + esc(bannerUrl ? "banner enviado pro cartao publico" : "banner do teu cartao publico") + "</p></article>" +
        "<article class='insight-card'><span>presenca</span><strong>" + esc(String(status || "online")) + "</strong><p>" + esc(bio ? "perfil com bio afinada" : "da pra deixar a bio mais viva") + "</p></article>";
    }
  }

  function refreshAppsNotesState(saved) {
    var notes = String(state.appsNotesDraft || "").trim();
    q("apps-notes-status").textContent = saved ? (notes ? "salvo local" : "limpo") : "falhou";
  }

  function handleGeneratePassword() {
    var length = Number(q("apps-password-length").value || 18);
    var password = generatePanelPassword(length);
    q("apps-password-output").value = password;
    q("apps-password-size").textContent = String(length) + " chars";
    toast("Senha forte gerada no Apps Lab.", "ok");
  }

  function handleSaveAppsNotes() {
    var saved = saveAppsNotes(state.appsNotesDraft);
    refreshAppsNotesState(saved);
    toast(saved ? "Notas locais salvas nesse dispositivo." : "Nao consegui salvar as notas locais.", saved ? "ok" : "err");
  }

  function renderMemberProfile() {
    var profile = state.selectedMember;
    var avatarWrap = q("member-avatar");
    var banner = q("member-banner");
    var sharedWrap = q("member-shared-rooms");
    var actionDM = q("member-action-dm");
    var actionBlock = q("member-action-block");
    var actionMute = q("member-action-mute");
    var relationshipWrap = q("member-relationship-pills");
    var moderationWrap = q("member-moderation");
    var roleSelect = q("member-role-select");
    var stats;
    var liveRoom;
    var statusCopy = "";
    if (!profile) {
      return;
    }
    stats = memberLocalStats(profile.user.userId || profile.user.id);
    liveRoom = profile.user.roomId ? roomById(profile.user.roomId) : (stats.lastRoom || null);
    statusCopy = personStatusCopy(profile.user);
    q("member-name").textContent = profile.user.displayName || profile.user.username || "Membro";
    q("member-meta").textContent = (profile.user.role || "member") + " // " + (profile.user.online ? (profile.user.status || "online") : "offline");
    q("member-status-pill").textContent = profile.user.online ? (profile.user.status || "online") : "offline";
    q("member-theme-pill").textContent = themeLabel(profile.user.theme || "matrix");
    q("member-bio").textContent = profile.user.bio || "Esse vivente ainda nao largou uma bio.";
    q("member-status-copy").textContent = statusCopy || "Sem status customizado.";
    q("member-status-copy").classList.toggle("hidden", !statusCopy);
    q("member-last-seen").textContent = profile.user.online ? "agora" : formatDateTime(profile.user.lastSeenAt);
    q("member-room-copy").textContent = liveRoom ? displayRoomName(liveRoom) : "Sem sala em foco";
    q("member-activity-copy").textContent = stats.lastMessage
      ? ("Ultimo rastro visivel em " + formatRelative(stats.lastMessage.createdAt))
      : "Sem atividade local carregada nesse dispositivo.";
    q("member-visible-messages").textContent = stats.messageCount + " msgs visiveis";
    q("member-visible-files").textContent = stats.attachmentCount + " midias visiveis";
    q("member-visible-rooms").textContent = stats.roomCount + " salas visiveis";
    q("member-relation-copy").textContent = profile.user.hasBlockedViewer
      ? "Esse usuario bloqueou teu contato."
      : (profile.user.blockedByViewer
        ? "Tu bloqueou esse usuario."
        : (profile.user.mutedByViewer ? "Tu silenciou esse usuario." : "Sem bloqueio ou silencio ativo."));
    if (relationshipWrap) {
      var sharedCount = sharedVisibleRoomsForUser(profile.user.userId || profile.user.id).length;
      relationshipWrap.innerHTML =
        "<span class='ghost-pill'>" + esc(profile.user.online ? "na base agora" : "offline") + "</span>" +
        "<span class='ghost-pill'>" + esc(sharedCount + " salas em comum") + "</span>" +
        "<span class='ghost-pill'>" + esc(profile.canDm ? "DM liberada" : "DM travada") + "</span>" +
        "<span class='ghost-pill'>" + esc(profile.user.hasBlockedViewer ? "te bloqueou" : (profile.user.blockedByViewer ? "bloqueado" : "contato livre")) + "</span>";
    }
    if (banner) {
      banner.style.background = panelBannerStyle(profile.user.theme || "matrix", profile.user.accentColor || themeAccent(profile.user.theme || "matrix"), profile.user.bannerPreset || "grid", profile.user.bannerUrl || "");
    }
    var memberAvatarUrl = safeAvatarUrl(profile.user.avatarUrl);
    if (memberAvatarUrl) {
      avatarWrap.innerHTML = "<img class='message-avatar profile-avatar-lg' alt='" + esc(profile.user.displayName || profile.user.username) + "' src='" + esc(memberAvatarUrl) + "' />";
    } else {
      avatarWrap.textContent = initials(profile.user.displayName || profile.user.username);
    }
    if (sharedWrap) {
      var sharedRooms = sharedVisibleRoomsForUser(profile.user.userId || profile.user.id);
      if (!sharedRooms.length) {
        sharedWrap.innerHTML = "<span class='member-shared-room more'>Sem sala compartilhada carregada ainda.</span>";
      } else {
        sharedWrap.innerHTML = sharedRooms.slice(0, 4).map(function(room) {
          var items = (state.messagesByRoom[String(room.id)] || []).filter(function(message) {
            return Number(message.authorId || message.userId || 0) === Number(profile.user.userId || profile.user.id);
          }).length;
          return "<button class='member-shared-room' type='button' data-action='jump-room' data-room-id='" + Number(room.id) + "'><strong>" + esc(displayRoomName(room)) + "</strong><span>" + items + " msgs</span></button>";
        }).join("") + (sharedRooms.length > 4 ? "<span class='member-shared-room more'>+" + (sharedRooms.length - 4) + " salas em comum</span>" : "");
      }
    }
    actionDM.textContent = profile.canManage ? "Editar meu perfil" : "Abrir DM";
    actionDM.disabled = !(profile.canManage || profile.canDm);
    actionBlock.textContent = profile.user.blockedByViewer ? "Desbloquear" : "Bloquear";
    actionMute.textContent = profile.user.mutedByViewer ? "Tirar silencio" : "Silenciar";
    actionBlock.disabled = profile.canManage;
    actionMute.disabled = profile.canManage;
    moderationWrap.classList.toggle("hidden", !profile.canModerate);
    if (profile.canModerate) {
      roleSelect.value = String(profile.user.role || "member");
      q("member-moderation-copy").textContent = "Dono online: mexe no cargo ou expulsa sem sair do perfil.";
    }
  }

  function fallbackSocialProfile(userId) {
    var presence = Number(userId) === Number(state.viewer && state.viewer.id)
      ? {
          userId: state.viewer.id,
          username: state.viewer.username,
          displayName: state.viewer.displayName,
          role: state.viewer.role,
          theme: state.viewer.theme,
          bannerPreset: state.viewer.bannerPreset,
          bannerUrl: state.viewer.bannerUrl,
          accentColor: state.viewer.accentColor,
          avatarUrl: state.viewer.avatarUrl,
          bio: state.viewer.bio,
          status: state.viewer.status,
          statusText: state.viewer.statusText,
          lastSeenAt: new Date().toISOString(),
          online: true,
          blockedByViewer: false,
          mutedByViewer: false,
          hasBlockedViewer: false
        }
      : presenceByUserId(userId);
    if (!presence) {
      return null;
    }
    return {
      user: presence,
      canDm: Number(userId) !== Number(state.viewer && state.viewer.id) && !presence.blockedByViewer && !presence.hasBlockedViewer && presence.role !== "ai",
      canManage: Number(userId) === Number(state.viewer && state.viewer.id),
      canModerate: state.viewer && state.viewer.role === "owner" && Number(userId) !== Number(state.viewer.id) && presence.role !== "owner" && presence.role !== "ai"
    };
  }

  function openGuideModal(force) {
    if (!force && hasSeenGuide()) {
      return;
    }
    q("guide-modal").classList.remove("hidden");
    syncTransientLayoutState();
  }

  function openAppCenterModal() {
    renderAppCenter();
    q("app-center-modal").classList.remove("hidden");
    syncTransientLayoutState();
  }

  function resetDownloadAccessState() {
    q("download-access-password").value = "";
    q("download-access-error").classList.add("hidden");
    q("download-access-error").textContent = "";
  }

  function resetJoinRequestState() {
    q("join-request-form").reset();
    q("join-request-error").classList.add("hidden");
    q("join-request-error").textContent = "";
  }

  function openJoinRequestModal() {
    resetJoinRequestState();
    q("join-request-modal").classList.remove("hidden");
    syncTransientLayoutState();
    q("join-request-email").focus();
  }

  function resetJoinCompleteState() {
    q("join-complete-form").reset();
    q("join-complete-error").classList.add("hidden");
    q("join-complete-error").textContent = "";
  }

  function openJoinCompleteModal() {
    resetJoinCompleteState();
    q("join-complete-modal").classList.remove("hidden");
    syncTransientLayoutState();
    q("join-complete-email").focus();
  }

  function openDownloadAccessModal() {
    resetDownloadAccessState();
    q("download-access-modal").classList.remove("hidden");
    syncTransientLayoutState();
    q("download-access-password").focus();
  }

  async function handleDownloadAccessSubmit(event) {
    event.preventDefault();
    var submit = q("btn-download-access-submit");
    var password = q("download-access-password").value.trim();
    if (!password) {
      q("download-access-error").classList.remove("hidden");
      q("download-access-error").textContent = "Bah, sem a senha do app nao tem como liberar esse download.";
      pulseClass(q("download-access-modal").querySelector(".modal-card"), "login-error-pulse");
      return;
    }
    try {
      setButtonBusy(submit, true, "Liberando...", "Liberar e baixar");
      var data = await apiFetch("/api/downloads/universald/access", {
        method: "POST",
        body: JSON.stringify({ password: password })
      });
      closeModal("download-access-modal");
      toast(data.message || "UniversalD liberado.", "ok");
      window.location.assign(data.downloadUrl || APP_DOWNLOAD_URL);
    } catch (err) {
      q("download-access-error").classList.remove("hidden");
      q("download-access-error").textContent = err.message || "Nao consegui liberar o download privado.";
      pulseClass(q("download-access-modal").querySelector(".modal-card"), "login-error-pulse");
      playTone("err");
    } finally {
      setButtonBusy(submit, false, "Liberando...", "Liberar e baixar");
    }
  }

  async function handleJoinRequestSubmit(event) {
    event.preventDefault();
    var submit = q("btn-join-request-submit");
    var card = q("join-request-modal").querySelector(".modal-card");
    var email = q("join-request-email").value.trim();
    var displayName = q("join-request-display").value.trim();
    var note = q("join-request-note").value.trim();
    if (!email || !displayName) {
      q("join-request-error").classList.remove("hidden");
      q("join-request-error").textContent = "Bah, manda email e nome antes de pedir guarida.";
      pulseClass(card, "login-error-pulse");
      return;
    }
    try {
      setButtonBusy(submit, true, "Enviando...", "Enviar pedido");
      var data = await apiFetch("/api/panel/join-request", {
        method: "POST",
        body: JSON.stringify({
          email: email,
          displayName: displayName,
          note: note
        })
      });
      closeModal("join-request-modal");
      q("login-error").classList.remove("hidden");
      q("login-error").textContent = data.message || "Pedido enviado. Se o dono aprovar, o codigo vai pro gmail ou sai manualmente por ele.";
      pulseClass(q("login-view").querySelector(".login-stage"), "login-success-pulse");
      playTone("ok");
    } catch (err) {
      q("join-request-error").classList.remove("hidden");
      q("join-request-error").textContent = err.message || "Nao consegui mandar teu pedido agora.";
      pulseClass(card, "login-error-pulse");
      playTone("err");
    } finally {
      setButtonBusy(submit, false, "Enviando...", "Enviar pedido");
    }
  }

  async function handleJoinCompleteSubmit(event) {
    event.preventDefault();
    var submit = q("btn-join-complete-submit");
    var card = q("join-complete-modal").querySelector(".modal-card");
    var payload = {
      email: q("join-complete-email").value.trim(),
      accessCode: q("join-complete-code").value.trim(),
      username: q("join-complete-username").value.trim(),
      displayName: q("join-complete-display").value.trim(),
      password: q("join-complete-password").value.trim()
    };
    if (!payload.email || !payload.accessCode || !payload.username || !payload.password) {
      q("join-complete-error").classList.remove("hidden");
      q("join-complete-error").textContent = "Sem email, codigo, usuario e senha essa conta nao nasce.";
      pulseClass(card, "login-error-pulse");
      return;
    }
    try {
      setButtonBusy(submit, true, "Criando...", "Criar acesso");
      var data = await apiFetch("/api/panel/join-request/complete", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      closeModal("join-complete-modal");
      q("login-input").value = payload.username;
      q("password-input").value = payload.password;
      q("login-error").classList.remove("hidden");
      q("login-error").textContent = data.message || "Acesso criado. Agora entra no painel com teu usuario novo.";
      pulseClass(q("login-view").querySelector(".login-stage"), "login-success-pulse");
      playTone("ok");
    } catch (err) {
      q("join-complete-error").classList.remove("hidden");
      q("join-complete-error").textContent = err.message || "Nao consegui concluir teu acesso.";
      pulseClass(card, "login-error-pulse");
      playTone("err");
    } finally {
      setButtonBusy(submit, false, "Criando...", "Criar acesso");
    }
  }

  function triggerUniversalDDownload() {
    if (state.viewer) {
      window.open(APP_DOWNLOAD_URL, "_blank", "noopener");
      toast("UniversalD saindo da base. Se quiser atualizar depois, volta no Apps Lab.", "ok");
      return;
    }
    openDownloadAccessModal();
  }

  function tryOpenUniversalD(options) {
    var opts = options || {};
    var shouldFallbackDownload = !!opts.fallbackToDownload;
    var source = String(opts.source || "painel-dief");
    var room = activeRoom();
    var protocolUrl = "universald://open?source=" + encodeURIComponent(source);
    if (state.activeNavId) {
      protocolUrl += "&nav=" + encodeURIComponent(state.activeNavId);
    }
    if (room && room.slug) {
      protocolUrl += "&room=" + encodeURIComponent(room.slug);
    }
    var cleared = false;

    function clearFallback() {
      if (cleared) {
        return;
      }
      cleared = true;
      window.clearTimeout(timer);
      window.removeEventListener("blur", clearFallback, true);
      document.removeEventListener("visibilitychange", handleVisibility, true);
    }

    function handleVisibility() {
      if (document.hidden) {
        saveAppInstallState(true);
        syncAppInstallStatus();
        renderAppCenter();
        clearFallback();
      }
    }

    var timer = window.setTimeout(function() {
      clearFallback();
      if (document.hidden) {
        return;
      }
      if (shouldFallbackDownload) {
        openDownloadAccessModal();
      } else {
        toast("Se o UniversalD nao abriu, baixa ou atualiza ele no Apps Lab.", "warn");
      }
    }, 1400);

    window.addEventListener("blur", clearFallback, true);
    document.addEventListener("visibilitychange", handleVisibility, true);
    toast("Tentando abrir o UniversalD nativo...", "ok");
    window.location.href = protocolUrl;
  }

  function handleOpenAppShortcut() {
    var room = roomBySlug("apps-lab");
    if (room && Number(state.activeRoomId) !== Number(room.id)) {
      selectRoom(room.id);
      window.setTimeout(function() {
        tryOpenUniversalD({ source: "apps-lab", fallbackToDownload: false });
      }, 160);
      return;
    }
    tryOpenUniversalD({ source: "painel-topbar", fallbackToDownload: false });
  }

  function maybeOpenGuide() {
    if (!state.viewer || hasSeenGuide()) {
      return;
    }
    window.setTimeout(function() {
      if (state.viewer) {
        openGuideModal(true);
      }
    }, 260);
  }

  async function openMemberProfile(userId) {
    try {
      var data = await apiFetch("/api/panel/social/profile?userId=" + encodeURIComponent(userId));
      state.selectedMember = data.profile;
      if (state.isMobile) {
        closeSidebar();
        closeInspector();
        state.utilityStripCollapsed = true;
        syncUtilityStripUI();
      }
      renderMemberProfile();
      q("member-modal").classList.remove("hidden");
      q("member-modal").scrollTop = 0;
      syncTransientLayoutState();
    } catch (err) {
      state.selectedMember = fallbackSocialProfile(userId);
      if (state.selectedMember) {
        if (state.isMobile) {
          closeSidebar();
          closeInspector();
          state.utilityStripCollapsed = true;
          syncUtilityStripUI();
        }
        renderMemberProfile();
        q("member-modal").classList.remove("hidden");
        q("member-modal").scrollTop = 0;
        syncTransientLayoutState();
        toast("Perfil abriu em modo local porque a API social oscilou.", "warn");
        return;
      }
      toast(err.message, "err");
    }
  }

  async function handleMemberDMAction() {
    if (!state.selectedMember) {
      return;
    }
    if (state.selectedMember.canManage) {
      closeModal("member-modal");
      openProfileModal();
      return;
    }
    if (!state.selectedMember.canDm) {
      toast("Essa conversa nao pode abrir nesse estado.", "warn");
      return;
    }
    await openDirectMessage(state.selectedMember.user.userId);
    closeModal("member-modal");
  }

  async function handleMemberBlockAction() {
    if (!state.selectedMember || state.selectedMember.canManage) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/social/block-toggle", {
        method: "POST",
        body: JSON.stringify({ targetUserId: Number(state.selectedMember.user.userId) })
      });
      syncLocalSocialFlags(state.selectedMember.user.userId, "blocked", !!data.blocked);
      state.selectedMember.user.blockedByViewer = !!data.blocked;
      state.selectedMember.canDm = !state.selectedMember.user.blockedByViewer && !state.selectedMember.user.hasBlockedViewer && state.selectedMember.user.role !== "ai";
      syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
      renderMemberProfile();
      if (state.activeRoomId) {
        await loadMessages(state.activeRoomId, true);
      }
      toast(data.blocked ? "Usuario bloqueado no teu painel." : "Bloqueio removido.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function handleMemberMuteAction() {
    if (!state.selectedMember || state.selectedMember.canManage) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/social/mute-toggle", {
        method: "POST",
        body: JSON.stringify({ targetUserId: Number(state.selectedMember.user.userId) })
      });
      syncLocalSocialFlags(state.selectedMember.user.userId, "muted", !!data.muted);
      state.selectedMember.user.mutedByViewer = !!data.muted;
      syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
      renderMemberProfile();
      if (state.activeRoomId) {
        await loadMessages(state.activeRoomId, true);
      }
      toast(data.muted ? "Usuario silenciado nas notificacoes." : "Silencio removido.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function handleMemberRoleSave() {
    if (!state.selectedMember || !state.selectedMember.canModerate) {
      return;
    }
    try {
      await updateUserRole(Number(state.selectedMember.user.userId), q("member-role-select").value);
      closeModal("member-modal");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function handleMemberExpelAction() {
    if (!state.selectedMember || !state.selectedMember.canModerate) {
      return;
    }
    if (!window.confirm("Quer mesmo expulsar esse usuario da base?")) {
      return;
    }
    try {
      await expelUser(Number(state.selectedMember.user.userId));
      closeModal("member-modal");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  function closeModal(id) {
    q(id).classList.add("hidden");
    if (id === "member-modal") {
      state.selectedMember = null;
    }
    if (id === "notify-modal") {
      state.notificationCenterOpen = false;
    }
    if (id === "media-modal") {
      state.mediaPreview = null;
    }
    if (id === "download-access-modal") {
      resetDownloadAccessState();
    }
    if (id === "join-request-modal") {
      resetJoinRequestState();
    }
    if (id === "join-complete-modal") {
      resetJoinCompleteState();
    }
    syncTransientLayoutState();
  }

  async function saveProfile(payload, successMessage) {
    try {
      var data = await apiFetch("/api/panel/profile", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      state.viewer = data.viewer;
      applyTheme();
      updateDocumentTitle();
      renderShell();
      toast(successMessage || "Perfil salvo.", "ok");
      return true;
    } catch (err) {
      toast(err.message, "err");
      return false;
    }
  }

  async function handleProfileSubmit(event) {
    event.preventDefault();
    var saved;
    setButtonBusy("btn-profile-submit", true, "Salvando...", "Salvar perfil");
    saved = await saveProfile({
      displayName: q("profile-display").value.trim(),
      bio: q("profile-bio").value.trim(),
      theme: q("profile-theme").value,
      bannerPreset: q("profile-banner-preset").value,
      bannerUrl: q("profile-banner-url").value.trim(),
      accentColor: q("profile-accent").value,
      avatarUrl: q("profile-avatar").value.trim(),
      status: q("profile-status").value,
      statusText: q("profile-status-text").value.trim()
    }, "Perfil salvo. Agora sim ta com tua cara.");
    setButtonBusy("btn-profile-submit", false, "Salvando...", "Salvar perfil");
    if (saved) {
      closeModal("profile-modal");
      sendPresence();
    }
  }

  async function handleCreateUser(event) {
    event.preventDefault();
    try {
      setButtonBusy("btn-create-user-submit", true, "Cadastrando...", "Cadastrar pessoa");
      await apiFetch("/api/panel/users", {
        method: "POST",
        body: JSON.stringify({
          username: q("create-user-name").value.trim(),
          displayName: q("create-user-display").value.trim(),
          email: q("create-user-email").value.trim(),
          password: q("create-user-password").value.trim(),
          role: q("create-user-role").value
        })
      });
      q("create-user-form").reset();
      await loadUsers();
      toast("Pessoa cadastrada pelo owner.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      setButtonBusy("btn-create-user-submit", false, "Cadastrando...", "Cadastrar pessoa");
    }
  }

  async function loadUsers() {
    if (!state.viewer || state.viewer.role !== "owner") {
      q("owner-card").classList.add("hidden");
      state.users = [];
      return;
    }
    try {
      var data = await apiFetch("/api/panel/users");
      state.users = data.users || [];
      renderUserRoster();
      q("owner-card").classList.remove("hidden");
    } catch (err) {
      state.users = [];
      renderUserRoster();
      q("owner-card").classList.add("hidden");
    }
  }

  async function loadJoinRequests(silent) {
    var previousPending = Number(state.pendingJoinCount || 0);
    if (!state.viewer || state.viewer.role !== "owner") {
      state.joinRequests = [];
      state.joinRequestsLoaded = false;
      state.pendingJoinCount = 0;
      renderJoinRequests();
      return;
    }
    try {
      var data = await apiFetch("/api/panel/join-requests");
      state.joinRequests = data.requests || [];
      state.joinRequestsLoaded = true;
      state.pendingJoinCount = state.joinRequests.filter(function(item) {
        return String(item.status || "") === "pending";
      }).length;
      renderJoinRequests();
      if (silent !== true && previousPending && state.pendingJoinCount > previousPending) {
        toast("Chegou pedido novo pra entrar na base.", "ok");
      }
    } catch (err) {
      state.joinRequests = [];
      renderJoinRequests();
      if (silent !== true) {
        toast(err.message, "err");
      }
    }
  }

  async function reviewJoinRequest(requestId, approve) {
    try {
      var data = await apiFetch("/api/panel/join-requests/review", {
        method: "POST",
        body: JSON.stringify({
          requestId: Number(requestId),
          approve: !!approve,
          reviewNote: ""
        })
      });
      syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
      await loadJoinRequests(true);
      if (approve) {
        toast((data.request && data.request.emailSent)
          ? "Pedido aprovado e codigo mandado pro gmail cadastrado."
          : "Pedido aprovado. O codigo ficou contigo pra passar manualmente.", "ok");
      } else {
        toast(data.message || "Pedido recusado com estilo debochado.", "warn");
      }
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function updateUserRole(userId, role) {
    var data = await apiFetch("/api/panel/users/role", {
      method: "POST",
      body: JSON.stringify({
        targetUserId: Number(userId),
        role: String(role || "member")
      })
    });
    await loadUsers();
    syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
    toast((data.user && data.user.displayName ? data.user.displayName : "Usuario") + " agora ta como " + (data.user && data.user.role ? data.user.role : role) + ".", "ok");
    return data;
  }

  async function expelUser(userId) {
    await apiFetch("/api/panel/users/expel", {
      method: "POST",
      body: JSON.stringify({ targetUserId: Number(userId) })
    });
    await loadUsers();
    syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
    toast("Usuario expulso da base.", "warn");
  }

  async function loadLogs() {
    if (!state.viewer || (state.viewer.role !== "owner" && state.viewer.role !== "admin")) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/logs?limit=80");
      state.recentLogs = data.logs || [];
      renderLogs();
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function openDirectMessage(userId) {
    try {
      var data = await apiFetch("/api/panel/dms/open", {
        method: "POST",
        body: JSON.stringify({ targetUserId: Number(userId) })
      });
      var refreshed = await apiFetch("/api/panel/bootstrap");
      syncBootstrapSnapshot(refreshed);
      await selectRoom(data.room.id, true);
      toast("DM aberta sem rodeio.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function runSearch(query) {
    var clean = String(query || "").trim();
    state.searchQuery = clean;
    state.searchResult = null;
    q("inspector-search-input").value = clean;
    setInspectorTab("search");
    if (!clean) {
      renderSearchResults();
      return;
    }
    try {
      setButtonBusy("btn-search-submit", true, "Buscando...", "Rodar busca");
      var data = await apiFetch("/api/panel/search?query=" + encodeURIComponent(clean) + "&limit=12");
      state.searchResult = data;
      renderSearchResults();
      toast("Busca rodada no painel.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      setButtonBusy("btn-search-submit", false, "Buscando...", "Rodar busca");
    }
  }

  function toggleActiveNavFavorite() {
    var navId = String(state.activeNavId || "");
    var wasFavorite = isFavoriteNavId(navId);
    if (!navId) {
      return;
    }
    state.favoriteNavIds = (state.favoriteNavIds || []).filter(function(item) {
      return String(item) !== navId;
    });
    var favorited = false;
    if (!wasFavorite) {
      state.favoriteNavIds.unshift(navId);
      favorited = true;
    }
    saveFavoriteNavIds();
    renderRooms();
    renderHeaderAndRoomState();
    toast(favorited ? "Area fixada no teu painel." : "Area removida dos favoritos.", "ok");
  }

  async function handleEventCreate(event) {
    event.preventDefault();
    var title = q("event-title").value.trim();
    var description = q("event-description").value.trim();
    var startsAtValue = q("event-starts-at").value;
    if (!title || !startsAtValue) {
      toast("Preenche titulo e horario do evento antes de soltar.", "warn");
      return;
    }
    try {
      setButtonBusy("btn-event-submit", true, "Criando...", "Criar evento");
      await apiFetch("/api/panel/events", {
        method: "POST",
        body: JSON.stringify({
          title: title,
          description: description,
          roomId: Number(q("event-room-id").value || 0),
          startsAt: new Date(startsAtValue).toISOString()
        })
      });
      q("event-form").reset();
      renderEventRoomOptions();
      syncBootstrapSnapshot(await apiFetch("/api/panel/bootstrap"));
      toast("Evento criado na agenda da base.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      setButtonBusy("btn-event-submit", false, "Criando...", "Criar evento");
    }
  }

  async function toggleEventRSVP(eventId) {
    try {
      var data = await apiFetch("/api/panel/events/rsvp-toggle", {
        method: "POST",
        body: JSON.stringify({ eventId: Number(eventId) })
      });
      state.events = (state.events || []).map(function(item) {
        return Number(item.id) === Number(data.event.id) ? data.event : item;
      });
      renderEvents();
      renderActivity();
      toast(data.joined ? "Presenca confirmada no evento." : "Tu saiu da lista do evento.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function deleteEvent(eventId) {
    var item = (state.events || []).find(function(event) {
      return Number(event.id) === Number(eventId);
    });
    if (!eventId || !item) {
      return;
    }
    if (!window.confirm("Apagar o evento \"" + (item.title || "sem titulo") + "\" da agenda?")) {
      return;
    }
    try {
      await apiFetch("/api/panel/events", {
        method: "DELETE",
        body: JSON.stringify({ eventId: Number(eventId) })
      });
      state.events = (state.events || []).filter(function(event) {
        return Number(event.id) !== Number(eventId);
      });
      renderEvents();
      renderActivity();
      toast("Evento removido da agenda.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function handlePollCreate(event) {
    var room = activeRoom();
    var question = q("poll-question").value.trim();
    var options = parsePollOptions(q("poll-options").value);
    event.preventDefault();
    if (!room || isHubView()) {
      toast("Escolhe uma sala antes de abrir enquete.", "warn");
      return;
    }
    if (!question || options.length < 2) {
      toast("A enquete precisa de pergunta e pelo menos duas opcoes.", "warn");
      return;
    }
    try {
      setButtonBusy("btn-poll-submit", true, "Criando...", "Criar enquete");
      await apiFetch("/api/panel/polls", {
        method: "POST",
        body: JSON.stringify({
          roomId: room.id,
          question: question,
          options: options
        })
      });
      q("poll-form").reset();
      await loadPolls(room.id, true);
      toast("Enquete criada nessa sala.", "ok");
    } catch (err) {
      toast(err.message, "err");
    } finally {
      setButtonBusy("btn-poll-submit", false, "Criando...", "Criar enquete");
    }
  }

  async function togglePollVote(pollId, optionId) {
    var room = activeRoom();
    if (!room || !pollId || !optionId) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/polls/vote-toggle", {
        method: "POST",
        body: JSON.stringify({
          roomId: room.id,
          pollId: Number(pollId),
          optionId: Number(optionId)
        })
      });
      state.polls = (state.polls || []).map(function(item) {
        return Number(item.id) === Number(data.poll.id) ? data.poll : item;
      });
      renderPolls();
      renderActivity();
      toast(data.voted ? "Voto salvo na enquete." : "Voto removido da enquete.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function deletePoll(pollId) {
    var room = activeRoom();
    var item = (state.polls || []).find(function(poll) {
      return Number(poll.id) === Number(pollId);
    });
    if (!room || !pollId || !item) {
      return;
    }
    if (!window.confirm("Apagar a enquete \"" + (item.question || "sem titulo") + "\" dessa sala?")) {
      return;
    }
    try {
      await apiFetch("/api/panel/polls", {
        method: "DELETE",
        body: JSON.stringify({
          roomId: Number(room.id),
          pollId: Number(pollId)
        })
      });
      state.polls = (state.polls || []).filter(function(poll) {
        return Number(poll.id) !== Number(pollId);
      });
      renderPolls();
      toast("Enquete removida da sala.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function runTerminal(event) {
    event.preventDefault();
    var command = q("terminal-input").value.trim();
    if (!command) {
      return;
    }
    q("terminal-output").textContent = "[rodando]\n" + command;
    try {
      setButtonBusy("btn-terminal-submit", true, "Rodando...", "Rodar");
      var data = await apiFetch("/api/panel/terminal/run", {
        method: "POST",
        body: JSON.stringify({ command: command })
      });
      var result = data.result || {};
      q("terminal-output").textContent =
        "[" + (result.exitCode || 0) + "] " + (result.command || command) + "\n\n" + (result.output || "(sem saida)");
      q("terminal-input").value = "";
      toast("Terminal respondeu da sala Apps Lab.", "ok");
    } catch (err) {
      q("terminal-output").textContent = "[erro]\n" + err.message;
      toast(err.message, "err");
    } finally {
      setButtonBusy("btn-terminal-submit", false, "Rodando...", "Rodar");
    }
  }

  async function toggleReaction(messageId, emoji) {
    var room = activeRoom();
    if (!room) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/reactions/toggle", {
        method: "POST",
        body: JSON.stringify({ roomId: room.id, messageId: messageId, emoji: emoji })
      });
      replaceRoomMessage(room.id, data.message);
      renderMessages(room.id);
      renderMoments();
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function sendPresence() {
    if (!state.viewer) {
      return;
    }
    try {
      await apiFetch("/api/panel/presence", {
        method: "POST",
        body: JSON.stringify({
          roomId: state.activeRoomId || 0,
          status: state.viewer.status || "online"
        })
      });
    } catch (err) {}
  }

  function setupHeartbeat() {
    if (state.heartbeatTimer) {
      window.clearInterval(state.heartbeatTimer);
    }
    state.heartbeatTimer = window.setInterval(function() {
      if (state.lastStreamAt && Date.now() - state.lastStreamAt > 30000) {
        setConnectionStatus("unstable");
      }
      sendPresence();
    }, 15000);
    sendPresence();
  }

  function compareOnlinePresence(nextOnline) {
    var nextMap = onlineMap(nextOnline);
    Object.keys(nextMap).forEach(function(userId) {
      if (String(state.viewer && state.viewer.id) === String(userId)) {
        return;
      }
      var person = nextOnline.find(function(item) { return String(item.userId) === String(userId); });
      if (!person || person.blockedByViewer || person.mutedByViewer) {
        return;
      }
      if (nextMap[userId] && !state.previousOnlineMap[userId] && person) {
        toast((person.displayName || person.username) + " colou no painel.", "ok");
        notifyBrowser("Painel Dief", (person.displayName || person.username) + " ficou online.");
        maybeVibrate([18, 40, 18]);
      }
      if (!nextMap[userId] && state.previousOnlineMap[userId] && person) {
        toast((person.displayName || person.username) + " vazou do painel.", "warn");
      }
    });
    state.previousOnlineMap = nextMap;
  }

  function syncRoomLatestChanges(nextLatestMap) {
    Object.keys(nextLatestMap).forEach(function(roomId) {
      var current = state.latestByRoom[roomId];
      var next = nextLatestMap[roomId];
      var currentId = Number(current && current.id || 0);
      var nextId = Number(next && next.id || 0);
      var room = roomById(roomId);
      if (!nextId || nextId <= currentId) {
        return;
      }
      if (String(roomId) !== String(state.activeRoomId) && Number(next.authorId) !== Number(state.viewer && state.viewer.id)) {
        if (next.blockedByViewer || isMutedUser(next.authorId)) {
          return;
        }
        state.unread[roomId] = Number(state.unread[roomId] || 0) + 1;
        state.latestUnreadByRoom[roomId] = Number(next.id);
        if (messageMentionsViewer(next)) {
          toast((next.authorName || "Alguem") + " te marcou em " + (room ? displayRoomName(room) : "uma sala") + ".", "warn");
          notifyBrowser("Painel Dief", (next.authorName || "Alguem") + " te marcou em " + (room ? displayRoomName(room) : "uma sala") + ".");
          maybeVibrate([26, 55, 26]);
        } else if (room && room.scope === "dm") {
          toast("Nova direta de " + (next.authorName || displayRoomName(room)) + ".", "ok");
          notifyBrowser("Painel Dief", "Nova DM de " + (next.authorName || displayRoomName(room)) + ".");
          maybeVibrate([22, 48, 22]);
        }
      }
    });
    state.latestByRoom = nextLatestMap;
    renderDashboardPulse();
    updateDocumentTitle();
  }

  function setupStream() {
    if (!window.EventSource) {
      setConnectionStatus("unstable");
      return;
    }
    if (state.stream) {
      state.stream.close();
    }
    setConnectionStatus("syncing");
    state.stream = new EventSource("/api/panel/stream");
    state.stream.addEventListener("snapshot", function(event) {
      try {
        handleSnapshot(JSON.parse(event.data));
      } catch (e) {}
    });
    state.stream.onerror = function() {
      setConnectionStatus("unstable");
      pushActivity("Stream oscilou, tentando reconectar sozinho.", "warn");
    };
  }

  function handleSnapshot(payload) {
    var nextRooms = payload.rooms || [];
    var nextLatestMap = latestMessagesMap(payload.latestMessages || []);
    var activeRoomId = String(state.activeRoomId || 0);
    var previousRoomVersion = Number(state.roomVersions[activeRoomId] || 0);
    var nextRoomVersions = payload.roomVersions || {};
    var nextActiveRoomVersion = Number(nextRoomVersions[activeRoomId] || previousRoomVersion || 0);
    var previousVersion = Number(state.version || 0);
    var nextVersion = Number(payload.version || state.version || 0);
    state.lastStreamAt = Date.now();
    setConnectionStatus("live");
    syncRoomLatestChanges(nextLatestMap);
    compareOnlinePresence(payload.online || []);

    state.rooms = nextRooms;
    state.roomAccess = payload.roomAccess || state.roomAccess;
    state.online = payload.online || [];
    state.typing = payload.typing || [];
    state.recentLogs = payload.recentLogs || state.recentLogs;
    state.events = payload.events || state.events;
    state.roomVersions = nextRoomVersions;
    state.blockedUserIds = (payload.blockedUserIds || state.blockedUserIds || []).map(Number);
    state.mutedUserIds = (payload.mutedUserIds || state.mutedUserIds || []).map(Number);
    state.version = nextVersion;

    renderShell();
    if (state.viewer && state.viewer.role === "owner" && nextVersion !== previousVersion) {
      loadUsers();
      loadJoinRequests(false);
    }

    var nextCurrentRoom = activeRoom();
    if (nextCurrentRoom && nextActiveRoomVersion !== previousRoomVersion && accessForRoom(nextCurrentRoom) !== "locked" && accessForRoom(nextCurrentRoom) !== "vip") {
      loadMessages(nextCurrentRoom.id, true);
      loadPolls(nextCurrentRoom.id, true);
    } else {
      renderTypingIndicator();
    }
  }

  async function sendTyping(active) {
    var room = activeRoom();
    if (!room || accessForRoom(room) === "locked" || accessForRoom(room) === "vip") {
      return;
    }
    try {
      await apiFetch("/api/panel/typing", {
        method: "POST",
        body: JSON.stringify({ roomId: room.id, active: active })
      });
    } catch (err) {}
  }

  function stopTyping(silent) {
    if (state.typingIdleTimer) {
      window.clearTimeout(state.typingIdleTimer);
      state.typingIdleTimer = null;
    }
    if (state.typingSent && !silent) {
      sendTyping(false);
    }
    state.typingSent = false;
  }

  function handleTypingInput() {
    var raw = q("composer-input").value || "";
    var value = raw.trim();
    persistComposerDraft();
    if (!value) {
      stopTyping(false);
      return;
    }
    if (!state.typingSent) {
      state.typingSent = true;
      sendTyping(true);
    }
    if (state.typingIdleTimer) {
      window.clearTimeout(state.typingIdleTimer);
    }
    state.typingIdleTimer = window.setTimeout(function() {
      stopTyping(false);
    }, 2200);
  }

  async function jumpToMessage(roomId, messageId) {
    delete state.latestUnreadByRoom[String(roomId)];
    await selectRoom(roomId, true, messageId);
    renderMessages(roomId);
  }

  function replaceRoomMessage(roomId, updatedMessage) {
    var roomKey = String(roomId);
    state.messagesByRoom[roomKey] = (state.messagesByRoom[roomKey] || []).map(function(item) {
      return Number(item.id) === Number(updatedMessage.id) ? updatedMessage : item;
    });
    state.pinnedByRoom[roomKey] = (state.pinnedByRoom[roomKey] || []).map(function(item) {
      return Number(item.id) === Number(updatedMessage.id) ? updatedMessage : item;
    });
    if (state.latestByRoom[roomKey] && Number(state.latestByRoom[roomKey].id) === Number(updatedMessage.id)) {
      state.latestByRoom[roomKey] = updatedMessage;
    }
  }

  function removeRoomMessage(roomId, messageId) {
    var roomKey = String(roomId);
    state.messagesByRoom[roomKey] = (state.messagesByRoom[roomKey] || []).filter(function(item) {
      return Number(item.id) !== Number(messageId);
    });
    state.pinnedByRoom[roomKey] = (state.pinnedByRoom[roomKey] || []).filter(function(item) {
      return Number(item.id) !== Number(messageId);
    });
    if (state.latestByRoom[roomKey] && Number(state.latestByRoom[roomKey].id) === Number(messageId)) {
      delete state.latestByRoom[roomKey];
    }
  }

  async function copyMessage(messageId) {
    var room = activeRoom();
    var items = state.messagesByRoom[String(room && room.id)] || [];
    var message = items.find(function(item) { return Number(item.id) === Number(messageId); });
    if (!message) {
      return;
    }
    await copyText(message.body || (message.attachment ? safeAttachmentUrl(message.attachment.url) : ""), "Mensagem copiada.", "Nao consegui copiar essa mensagem.");
  }

  function startReply(messageId) {
    var room = activeRoom();
    var items = state.messagesByRoom[String(room && room.id)] || [];
    var message = items.find(function(item) { return Number(item.id) === Number(messageId); });
    if (!message) {
      return;
    }
    state.editingMessage = null;
    state.replyTarget = {
      id: message.id,
      authorName: message.authorName,
      body: message.body || (message.attachment ? "[anexo] " + message.attachment.name : "")
    };
    renderReplyChip();
    renderEditChip();
    renderHeaderAndRoomState();
    syncComposerDraftHint();
    q("composer-input").focus();
  }

  function startEdit(messageId) {
    var room = activeRoom();
    var items = state.messagesByRoom[String(room && room.id)] || [];
    var message = items.find(function(item) { return Number(item.id) === Number(messageId); });
    if (!message || !canManageMessage(message)) {
      return;
    }
    state.replyTarget = null;
    state.pendingUploads = [];
    state.pendingAttachment = null;
    state.editingMessage = {
      id: message.id,
      authorName: message.authorName,
      attachment: message.attachment || null,
      hasAttachment: !!message.attachment
    };
    q("composer-input").value = message.body || "";
    renderPendingAttachment();
    renderReplyChip();
    renderEditChip();
    renderHeaderAndRoomState();
    syncComposerDraftHint();
    q("composer-input").focus();
  }

  async function deleteMessage(messageId) {
    var room = activeRoom();
    if (!room) {
      return;
    }
    if (!window.confirm("Apagar essa mensagem de vez do Painel Dief?")) {
      return;
    }
    try {
      await apiFetch("/api/panel/messages", {
        method: "DELETE",
        body: JSON.stringify({ roomId: room.id, messageId: Number(messageId) })
      });
      removeRoomMessage(room.id, messageId);
      if (state.editingMessage && Number(state.editingMessage.id) === Number(messageId)) {
        state.editingMessage = null;
      }
      if (state.replyTarget && Number(state.replyTarget.id) === Number(messageId)) {
        state.replyTarget = null;
      }
      renderReplyChip();
      renderEditChip();
      renderHeaderAndRoomState();
      renderRooms();
      renderMessages(room.id);
      renderPinnedStrip();
      renderFavorites();
      renderMoments();
      toast("Mensagem apagada do mapa.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function togglePin(messageId) {
    var room = activeRoom();
    if (!room) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/pins/toggle", {
        method: "POST",
        body: JSON.stringify({ roomId: room.id, messageId: Number(messageId) })
      });
      replaceRoomMessage(room.id, data.message);
      if (data.pinned) {
        state.pinnedByRoom[String(room.id)] = [data.message].concat((state.pinnedByRoom[String(room.id)] || []).filter(function(item) {
          return Number(item.id) !== Number(messageId);
        })).slice(0, 6);
      } else {
        state.pinnedByRoom[String(room.id)] = (state.pinnedByRoom[String(room.id)] || []).filter(function(item) {
          return Number(item.id) !== Number(messageId);
        });
      }
      renderMessages(room.id);
      renderPinnedStrip();
      renderMoments();
      toast(data.pinned ? "Mensagem fixada no topo." : "Mensagem desfixada.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  async function toggleFavorite(messageId) {
    var room = activeRoom();
    if (!room) {
      return;
    }
    try {
      var data = await apiFetch("/api/panel/favorites/toggle", {
        method: "POST",
        body: JSON.stringify({ roomId: room.id, messageId: Number(messageId) })
      });
      replaceRoomMessage(room.id, data.message);
      renderMessages(room.id);
      renderFavorites();
      renderMoments();
      toast(data.favorited ? "Favorito salvo pra achar rapido." : "Favorito removido.", "ok");
    } catch (err) {
      toast(err.message, "err");
    }
  }

  function setInspectorTab(tab) {
    var map = {
      overview: "inspector-section-overview",
      members: "inspector-section-members",
      files: "inspector-section-files",
      search: "inspector-section-search",
      logs: "inspector-section-logs",
      admin: "inspector-section-admin"
    };
    var sectionId = "";
    state.inspectorTab = tab;
    renderInspectorTabs();
    openInspector();
    sectionId = map[tab];
    if (tab === "logs") {
      loadLogs();
    }
    if (tab === "search") {
      window.setTimeout(function() {
        if (q("inspector-search-input")) {
          q("inspector-search-input").focus();
        }
      }, 40);
    }
    if (sectionId && q(sectionId) && q(sectionId).scrollIntoView) {
      window.setTimeout(function() {
        safeScrollIntoView(q(sectionId), { block: "start", behavior: "smooth" });
      }, 80);
    }
  }

  function startMatrixBackground() {
    var canvas = q("matrix-bg");
    if (!canvas || !canvas.getContext) {
      return;
    }
    var ctx = canvas.getContext("2d");
    var columns = [];
    var glyphs = "01PDNEGOdramias<>[]{}#$%";
    var fontSize = 16;

    function resize() {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
      var count = Math.ceil(canvas.width / fontSize);
      columns = [];
      for (var i = 0; i < count; i++) {
        columns.push(Math.random() * canvas.height / fontSize);
      }
    }

    function draw() {
      ctx.fillStyle = "rgba(2, 6, 4, 0.12)";
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.font = "700 " + fontSize + "px JetBrains Mono";
      var accent = getComputedStyle(document.documentElement).getPropertyValue("--accent") || "#7bff00";
      for (var i = 0; i < columns.length; i++) {
        var text = glyphs.charAt(Math.floor(Math.random() * glyphs.length));
        var x = i * fontSize;
        var y = columns[i] * fontSize;
        ctx.fillStyle = accent;
        ctx.fillText(text, x, y);
        if (y > canvas.height && Math.random() > 0.975) {
          columns[i] = 0;
        }
        columns[i] += 0.95 + Math.random() * 0.55;
      }
      window.requestAnimationFrame(draw);
    }

    resize();
    window.addEventListener("resize", resize);
    draw();
  }

  function bindEvents() {
    syncViewportMetrics();
    detectMobile();
    window.addEventListener("resize", scheduleLayoutRefresh);
    if (window.visualViewport) {
      window.visualViewport.addEventListener("resize", scheduleLayoutRefresh);
      window.visualViewport.addEventListener("scroll", scheduleLayoutRefresh);
    }
    document.addEventListener("focusin", function(event) {
      var target = event.target;
      var tag = String(target && target.tagName || "").toLowerCase();
      if (tag === "input" || tag === "textarea" || tag === "select" || !!(target && target.isContentEditable)) {
        ensureFieldVisible(target);
      }
    });
    q("login-form").addEventListener("submit", handleLogin);
    q("btn-login-open-app").addEventListener("click", function() {
      tryOpenUniversalD({ source: "login", fallbackToDownload: true });
    });
    q("btn-login-download").addEventListener("click", openDownloadAccessModal);
    q("btn-login-guide").addEventListener("click", openAppCenterModal);
    q("btn-login-request-access").addEventListener("click", openJoinRequestModal);
    q("btn-login-complete-access").addEventListener("click", openJoinCompleteModal);
    q("btn-logout").addEventListener("click", handleLogout);
    q("btn-profile").addEventListener("click", openProfileModal);
    q("btn-logout-utility").addEventListener("click", handleLogout);
    q("btn-profile-utility").addEventListener("click", openProfileModal);
    q("btn-open-app").addEventListener("click", handleOpenAppShortcut);
    q("btn-guide").addEventListener("click", function() { openGuideModal(true); });
    q("btn-notify-center").addEventListener("click", openNotificationCenter);
    q("btn-notify-clear").addEventListener("click", clearNotificationCenter);
    q("btn-focus-mode").addEventListener("click", toggleFocusMode);
    q("btn-context-toggle").addEventListener("click", toggleChatContext);
    q("btn-utility-toggle").addEventListener("click", toggleUtilityStrip);
    q("btn-browser-notify").addEventListener("click", handleBrowserNotifyToggle);
    if (q("btn-app-open")) {
      q("btn-app-open").addEventListener("click", function() {
        tryOpenUniversalD({ source: "apps-lab", fallbackToDownload: false });
      });
    }
    if (q("btn-app-download")) {
      q("btn-app-download").addEventListener("click", triggerUniversalDDownload);
    }
    if (q("btn-app-guide")) {
      q("btn-app-guide").addEventListener("click", openAppCenterModal);
    }
    q("btn-app-center-open").addEventListener("click", function() {
      tryOpenUniversalD({ source: "app-center", fallbackToDownload: true });
    });
    q("btn-app-center-download").addEventListener("click", triggerUniversalDDownload);
    q("btn-app-center-copy").addEventListener("click", function() {
      copyText(window.location.origin, "Link da base copiado.", "Nao consegui copiar o link da base.");
    });
    q("btn-app-center-guide").addEventListener("click", function() {
      closeModal("app-center-modal");
      openGuideModal(true);
    });
    q("download-access-form").addEventListener("submit", handleDownloadAccessSubmit);
    q("join-request-form").addEventListener("submit", handleJoinRequestSubmit);
    q("join-complete-form").addEventListener("submit", handleJoinCompleteSubmit);
    q("btn-room-favorite").addEventListener("click", toggleActiveNavFavorite);
    q("member-action-dm").addEventListener("click", handleMemberDMAction);
    q("member-action-block").addEventListener("click", handleMemberBlockAction);
    q("member-action-mute").addEventListener("click", handleMemberMuteAction);
    q("member-action-role-save").addEventListener("click", handleMemberRoleSave);
    q("member-action-expel").addEventListener("click", handleMemberExpelAction);
    q("btn-audio-toggle").addEventListener("click", toggleAudio);
    q("btn-sidebar-open").addEventListener("click", openSidebar);
    q("composer-form").addEventListener("submit", handleComposerSubmit);
    q("composer-input").addEventListener("input", handleTypingInput);
    q("composer-input").addEventListener("focus", function() {
      setComposerFocusMode(true);
      if (state.compactLayout) {
        state.utilityStripCollapsed = true;
        syncUtilityStripUI();
        if (!isChatContextForcedOpen() && !isSusView()) {
          state.chatContextCollapsed = true;
          applyChatContextState();
        }
      }
      ensureFieldVisible(q("composer-input"));
      window.setTimeout(function() {
        ensureFieldVisible(q("composer-form"));
      }, 120);
    });
    q("composer-input").addEventListener("blur", function() {
      window.setTimeout(function() {
        setComposerFocusMode(false);
      }, 120);
    });
    q("composer-input").addEventListener("keydown", function(event) {
      if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault();
        submitComposer();
      }
    });
    q("btn-ai").addEventListener("click", function() {
      submitComposer();
    });
    q("room-filter-input").addEventListener("input", function(event) {
      state.roomSearch = event.target.value || "";
      renderRooms();
    });
    q("btn-attach").addEventListener("click", function() { q("attachment-input").click(); });
    q("attachment-input").addEventListener("change", async function() {
      try {
        await queueFilesForUpload(q("attachment-input").files, { mode: "composer" });
        toast("Fila de upload atualizada.", "ok");
      } catch (err) {
        toast(err.message, "err");
      } finally {
        q("attachment-input").value = "";
      }
    });
    q("composer-form").addEventListener("dragenter", function(event) {
      if (state.isMobile) {
        return;
      }
      event.preventDefault();
      q("composer-form").classList.add("drag-over");
    });
    q("composer-form").addEventListener("dragover", function(event) {
      if (state.isMobile) {
        return;
      }
      event.preventDefault();
      q("composer-form").classList.add("drag-over");
    });
    q("composer-form").addEventListener("dragleave", function(event) {
      if (state.isMobile) {
        return;
      }
      if (!q("composer-form").contains(event.relatedTarget)) {
        q("composer-form").classList.remove("drag-over");
      }
    });
    q("composer-form").addEventListener("drop", async function(event) {
      if (state.isMobile) {
        return;
      }
      event.preventDefault();
      q("composer-form").classList.remove("drag-over");
      if (!event.dataTransfer || !event.dataTransfer.files || !event.dataTransfer.files.length) {
        return;
      }
      try {
        await queueFilesForUpload(event.dataTransfer.files, { mode: "composer" });
        toast("Arquivos encaixados na fila do chat.", "ok");
      } catch (err) {
        toast(err.message, "err");
      }
    });
    q("message-stream").addEventListener("scroll", function() {
      collapseMobileChrome(false);
    }, { passive: true });
    q("message-stream").addEventListener("touchstart", function() {
      collapseMobileChrome(true);
    }, { passive: true });
    q("conversation-panel").addEventListener("touchstart", function() {
      collapseMobileChrome(false);
    }, { passive: true });
    q("unlock-form").addEventListener("submit", handleUnlockSubmit);
    q("btn-room-action").addEventListener("click", function() {
      var room = activeRoom();
      if (room && accessForRoom(room) === "locked") {
        openUnlockModal(room.id);
      } else if (room && room.scope === "dm" && directPeerProfile(room) && (directPeerProfile(room).blockedByViewer || directPeerProfile(room).hasBlockedViewer)) {
        openMemberProfile(directPeerProfile(room).userId);
      } else {
        toast("Sem acesso liberado nessa aba.", "warn");
      }
    });
    q("profile-form").addEventListener("submit", handleProfileSubmit);
    q("profile-theme").addEventListener("change", function() {
      syncThemePresetState();
      renderProfileFormPreview();
    });
    q("profile-banner-preset").addEventListener("change", function() {
      syncBannerPresetState();
      renderProfileFormPreview();
    });
    Array.prototype.slice.call(document.querySelectorAll("[data-theme-preset]")).forEach(function(button) {
      button.addEventListener("click", function() {
        applyThemePreset(button.getAttribute("data-theme-preset"), button.getAttribute("data-theme-accent"));
        renderProfileFormPreview();
      });
    });
    Array.prototype.slice.call(document.querySelectorAll("[data-banner-preset]")).forEach(function(button) {
      button.addEventListener("click", function() {
        applyBannerPreset(button.getAttribute("data-banner-preset"));
        renderProfileFormPreview();
      });
    });
    q("btn-upload-avatar").addEventListener("click", function() { q("avatar-upload-input").click(); });
    q("avatar-upload-input").addEventListener("change", function() {
      uploadSelectedFile(q("avatar-upload-input"), function(attachment) {
        q("profile-avatar").value = safeAttachmentUrl(attachment.url);
        renderProfileFormPreview();
      });
    });
    q("btn-upload-banner").addEventListener("click", function() { q("banner-upload-input").click(); });
    q("btn-clear-banner").addEventListener("click", function() {
      q("profile-banner-url").value = "";
      renderProfileFormPreview();
    });
    q("banner-upload-input").addEventListener("change", function() {
      uploadSelectedFile(q("banner-upload-input"), function(attachment) {
        q("profile-banner-url").value = safeAttachmentUrl(attachment.url);
        renderProfileFormPreview();
      });
    });
    ["profile-display", "profile-status", "profile-accent", "profile-avatar", "profile-bio", "profile-status-text", "profile-banner-preset", "profile-banner-url"].forEach(function(id) {
      q(id).addEventListener("input", renderProfileFormPreview);
      q(id).addEventListener("change", renderProfileFormPreview);
    });
    q("create-user-form").addEventListener("submit", handleCreateUser);
    q("event-form").addEventListener("submit", handleEventCreate);
    q("poll-form").addEventListener("submit", handlePollCreate);
    q("terminal-form").addEventListener("submit", runTerminal);
    q("btn-guide-dismiss").addEventListener("click", function() {
      markGuideSeen();
      closeModal("guide-modal");
      toast("Guia fechado. Se quiser rever, usa o botao Guia.", "ok");
    });
    q("btn-guide-remind").addEventListener("click", function() {
      clearGuideSeen();
      closeModal("guide-modal");
      toast("Beleza. O guia fica disponivel quando tu quiser.", "ok");
    });
    q("search-form").addEventListener("submit", function(event) {
      event.preventDefault();
      runSearch(q("inspector-search-input").value.trim());
    });
    q("members-search-input").addEventListener("input", function(event) {
      state.memberSearch = event.target.value || "";
      renderPresenceSidebar();
    });
    q("members-role-filter").addEventListener("change", function(event) {
      state.memberRoleFilter = event.target.value || "all";
      renderPresenceSidebar();
    });
    q("media-search-input").addEventListener("input", function(event) {
      state.mediaSearch = event.target.value || "";
      renderFiles();
      renderMediaPreview();
    });
    q("media-sort-select").addEventListener("change", function(event) {
      state.mediaSort = event.target.value || "recent";
      renderFiles();
      renderMediaPreview();
    });
    Array.prototype.slice.call(document.querySelectorAll("[data-media-filter]")).forEach(function(button) {
      button.addEventListener("click", function() {
        state.mediaFilter = button.getAttribute("data-media-filter") || "all";
        Array.prototype.slice.call(document.querySelectorAll("[data-media-filter]")).forEach(function(item) {
          item.classList.toggle("active", item === button);
        });
        renderFiles();
        renderMediaPreview();
      });
    });
    q("media-modal-prev").addEventListener("click", function() { stepMediaPreview(-1); });
    q("media-modal-next").addEventListener("click", function() { stepMediaPreview(1); });
    q("media-modal-origin").addEventListener("click", function() {
      var roomId = Number(q("media-modal-origin").getAttribute("data-room-id") || 0);
      var messageId = Number(q("media-modal-origin").getAttribute("data-message-id") || 0);
      if (!roomId || !messageId) {
        return;
      }
      closeModal("media-modal");
      jumpToMessage(roomId, messageId);
    });
    q("btn-refresh-logs").addEventListener("click", loadLogs);
    q("btn-sidebar-peek").addEventListener("click", openSidebar);
    q("btn-sidebar-close").addEventListener("click", closeSidebar);
    q("btn-inspector-peek").addEventListener("click", openInspector);
    q("btn-inspector-toggle").addEventListener("click", function() {
      setInspectorTab("overview");
    });
    q("btn-inspector-close").addEventListener("click", closeInspector);
    q("mobile-backdrop").addEventListener("click", function() {
      closeSidebar();
      closeInspector();
    });
    window.addEventListener("keydown", function(event) {
      var tag = String(event.target && event.target.tagName || "").toLowerCase();
      var typingField = tag === "input" || tag === "textarea" || tag === "select" || !!(event.target && event.target.isContentEditable);
      if ((event.ctrlKey || event.metaKey) && String(event.key || "").toLowerCase() === "k") {
        event.preventDefault();
        setInspectorTab("search");
        return;
      }
      if ((event.ctrlKey || event.metaKey) && String(event.key || "") === ".") {
        if (!typingField) {
          event.preventDefault();
          q("composer-input").focus();
        }
        return;
      }
      if (!typingField && !event.ctrlKey && !event.metaKey && !event.altKey && event.key === "/") {
        event.preventDefault();
        setInspectorTab("search");
        q("inspector-search-input").focus();
        q("inspector-search-input").select();
        return;
      }
      if (event.key === "Escape") {
        closeModal("profile-modal");
        closeModal("unlock-modal");
        closeModal("media-modal");
        closeModal("guide-modal");
        closeModal("notify-modal");
        closeModal("app-center-modal");
        closeModal("download-access-modal");
        closeModal("join-request-modal");
        closeModal("join-complete-modal");
        closeSidebar();
        closeInspector();
      }
    });

    document.body.addEventListener("click", function(event) {
      var target = event.target;
      if (!target) {
        return;
      }
      var closeId = target.getAttribute && target.getAttribute("data-close-modal");
      if (closeId) {
        closeModal(closeId);
      }
      if (target === q("profile-modal")) {
        closeModal("profile-modal");
      }
      if (target === q("member-modal")) {
        closeModal("member-modal");
      }
      if (target === q("media-modal")) {
        closeModal("media-modal");
      }
      if (target === q("unlock-modal")) {
        closeModal("unlock-modal");
      }
      if (target === q("guide-modal")) {
        closeModal("guide-modal");
      }
      if (target === q("notify-modal")) {
        closeModal("notify-modal");
      }
      if (target === q("app-center-modal")) {
        closeModal("app-center-modal");
      }
      if (target === q("download-access-modal")) {
        closeModal("download-access-modal");
      }
      if (target === q("join-request-modal")) {
        closeModal("join-request-modal");
      }
      if (target === q("join-complete-modal")) {
        closeModal("join-complete-modal");
      }
    });

    document.body.addEventListener("click", function(event) {
      var rawTarget = event.target;
      var target = rawTarget && rawTarget.closest ? rawTarget.closest("[data-room-id], [data-nav-id], [data-action]") : null;
      if (!target) {
        if (rawTarget && rawTarget.id === "btn-clear-reply") {
          state.replyTarget = null;
          renderReplyChip();
        }
        if (rawTarget && rawTarget.id === "btn-clear-edit") {
          state.editingMessage = null;
          renderEditChip();
          renderHeaderAndRoomState();
        }
        return;
      }

      if (target.hasAttribute("data-nav-id")) {
        blurInteractiveTarget(target);
        selectPrimaryNav(target.getAttribute("data-nav-id"));
        return;
      }
      if (target.hasAttribute("data-room-id") && target.classList.contains("room-item")) {
        blurInteractiveTarget(target);
        selectRoom(Number(target.getAttribute("data-room-id")));
        return;
      }

      var action = target.getAttribute("data-action");
      if (!action) {
        return;
      }
      if (action === "reply") {
        startReply(Number(target.getAttribute("data-message-id")));
      } else if (action === "open-user-profile") {
        openMemberProfile(Number(target.getAttribute("data-user-id")));
      } else if (action === "edit") {
        startEdit(Number(target.getAttribute("data-message-id")));
      } else if (action === "delete") {
        deleteMessage(Number(target.getAttribute("data-message-id")));
      } else if (action === "copy") {
        copyMessage(Number(target.getAttribute("data-message-id")));
      } else if (action === "react") {
        toggleReaction(Number(target.getAttribute("data-message-id")), target.getAttribute("data-emoji"));
      } else if (action === "pin") {
        togglePin(Number(target.getAttribute("data-message-id")));
      } else if (action === "favorite") {
        toggleFavorite(Number(target.getAttribute("data-message-id")));
      } else if (action === "open-dm") {
        openDirectMessage(Number(target.getAttribute("data-user-id")));
      } else if (action === "preview-attachment") {
        openMediaPreview(Number(target.getAttribute("data-room-id")), Number(target.getAttribute("data-message-id")));
      } else if (action === "remove-upload") {
        removePendingUpload(target.getAttribute("data-upload-id"));
      } else if (action === "retry-upload") {
        retryPendingUpload(target.getAttribute("data-upload-id"));
      } else if (action === "open-inspector-section") {
        setInspectorTab(target.getAttribute("data-section") || "overview");
      } else if (action === "jump-room") {
        selectRoom(Number(target.getAttribute("data-room-id")));
      } else if (action === "jump-message") {
        jumpToMessage(Number(target.getAttribute("data-room-id")), Number(target.getAttribute("data-message-id")));
      } else if (action === "notification-jump") {
        closeModal("notify-modal");
        if (target.getAttribute("data-message-id")) {
          jumpToMessage(Number(target.getAttribute("data-room-id")), Number(target.getAttribute("data-message-id")));
        } else {
          selectRoom(Number(target.getAttribute("data-room-id")));
        }
      } else if (action === "jump-unread") {
        jumpToMessage(Number(target.getAttribute("data-room-id")), Number(target.getAttribute("data-message-id")));
      } else if (action === "focus-composer") {
        if (q("composer-form").classList.contains("hidden")) {
          toast("Abre uma conversa liberada antes de responder, vivente.", "warn");
        } else {
          q("composer-input").focus();
          ensureFieldVisible(q("composer-input"));
        }
      } else if (action === "trigger-attach") {
        if (q("composer-form").classList.contains("hidden")) {
          toast("Essa area nao aceita anexo agora. Abre uma sala liberada primeiro.", "warn");
        } else {
          q("attachment-input").click();
        }
      } else if (action === "open-app-center") {
        openAppCenterModal();
      } else if (action === "toggle-event-rsvp") {
        toggleEventRSVP(Number(target.getAttribute("data-event-id")));
      } else if (action === "vote-poll") {
        togglePollVote(Number(target.getAttribute("data-poll-id")), Number(target.getAttribute("data-option-id")));
      } else if (action === "delete-event") {
        deleteEvent(Number(target.getAttribute("data-event-id")));
      } else if (action === "delete-poll") {
        deletePoll(Number(target.getAttribute("data-poll-id")));
      } else if (action === "sus-trigger") {
        handleSusTrigger();
      } else if (action === "review-join-request") {
        reviewJoinRequest(Number(target.getAttribute("data-request-id")), target.getAttribute("data-approve") === "true");
      } else if (action === "apply-user-role") {
        var roleInput = document.querySelector("[data-role-select='" + Number(target.getAttribute("data-user-id")) + "']");
        updateUserRole(Number(target.getAttribute("data-user-id")), roleInput ? roleInput.value : "member").catch(function(err) {
          toast(err.message, "err");
        });
      } else if (action === "expel-user") {
        if (window.confirm("Quer expulsar esse usuario da base?")) {
          expelUser(Number(target.getAttribute("data-user-id"))).catch(function(err) {
            toast(err.message, "err");
          });
        }
      }
    });
  }

  function init() {
    loadAppInstallState();
    startMatrixBackground();
    renderAppCenter();
    syncAppInstallStatus();
    renderConnectionStatus();
    bindEvents();
    tryBootstrap();
  }

  window.addEventListener("error", function() {
    enterEmergencyMode("Bah, teu navegador entrou em modo de compatibilidade. Tenta entrar de novo.");
  });
  window.addEventListener("unhandledrejection", function() {
    enterEmergencyMode("Bah, deu uma oscilada no mobile. Ativei o modo de compatibilidade.");
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function() {
      try {
        init();
      } catch (err) {
        enterEmergencyMode("Modo de compatibilidade ativado no carregamento. Tenta entrar de novo.");
      }
    });
  } else {
    try {
      init();
    } catch (err) {
      enterEmergencyMode("Modo de compatibilidade ativado no carregamento. Tenta entrar de novo.");
    }
  }
})();
