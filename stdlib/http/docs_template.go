package http

const htmlDocsTemplate = `<!DOCTYPE html>
<html lang="pt-BR" class="min-h-full tema-escuro">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Harpia API Console 🦅</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style id="harpia-vars">
        html.tema-escuro, html.tema-escuro body {
            color-scheme: dark !important;
            background-color: #0A0F0D !important;
            color: #F3F6F4 !important;
            min-height: 100% !important;
        }
        html.tema-escuro header {
            background-color: transparent !important;
            border-bottom: none !important;
            color: #F3F6F4 !important;
        }
        html.tema-escuro #header-pill {
            background-color: rgba(23, 30, 38, 0.95) !important;
            border: 1px solid rgba(255, 255, 255, 0.08) !important;
            border-radius: 9999px !important;
            backdrop-filter: blur(12px) !important;
            -webkit-backdrop-filter: blur(12px) !important;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5) !important;
        }
        html.tema-escuro #barra-filtros-wrapper {
            background-color: #0A0F0D !important;
        }
        html.tema-escuro #barra-filtros {
            background-color: #171E26 !important;
            border: 1px solid rgba(255, 255, 255, 0.08) !important;
            border-radius: 9999px !important;
            padding: 0.375rem 0.875rem !important;
            width: fit-content !important;
            display: inline-flex !important;
            flex-wrap: wrap !important;
            align-items: center !important;
        }
        body { font-family: 'Inter', sans-serif; background-color: #0A0F0D; color: #F3F6F4; }
        .font-mono-harpia { font-family: 'JetBrains Mono', monospace; }
        ::-webkit-scrollbar { width: 6px; height: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: rgba(140, 155, 158, 0.3); border-radius: 3px; }
        ::-webkit-scrollbar-thumb:hover { background: rgba(140, 155, 158, 0.5); }
        .has-tip { cursor: help; }
        #tooltip-float {
            position: fixed; background: #171E26; border: 1px solid #F2A900;
            color: #F3F6F4; font-size: 10px; font-weight: 400; padding: 6px 10px;
            border-radius: 6px; white-space: nowrap; pointer-events: none;
            z-index: 2147483647; box-shadow: 0 4px 12px rgba(0,0,0,0.5);
            font-family: 'Inter', sans-serif; opacity: 0; transition: opacity 0.12s;
        }
        @keyframes pulse-verde { 0%, 100% { opacity: 1; transform: scale(1); } 50% { opacity: 0.4; transform: scale(0.92); } }
        @keyframes pulse-vermelho { 0%, 100% { opacity: 1; } 50% { opacity: 0.2; } }
        .pulse-online { animation: pulse-verde 2s infinite ease-in-out; }
        .pulse-offline { animation: pulse-vermelho 1s infinite ease-in-out; }
        .ativa-rota { color: #F2A900 !important; border-left: 2px solid #F2A900; background-color: #171E26; }
        @keyframes flash-verde { 0%, 100% { background-color: transparent; } 50% { background-color: rgba(0, 168, 107, 0.25); } }
        .flash-mudanca { animation: flash-verde 0.8s ease-in-out; }

        /* Animação Suave de Slide Up & Down do Cabeçalho Pílula */
        header {
            will-change: transform, opacity;
            transform: translateY(0);
            opacity: 1;
            transition: transform 0.6s cubic-bezier(0.34, 1.35, 0.64, 1), opacity 0.45s ease-out !important;
        }
        #header-pill {
            transition: padding 0.35s cubic-bezier(0.16, 1, 0.3, 1), background-color 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease !important;
        }

        /* Estado compacto (shrink ao rolar a página) */
        header.header-shrunk #header-pill {
            padding-top: 0.35rem !important;
            padding-bottom: 0.35rem !important;
            padding-left: 1rem !important;
            padding-right: 1rem !important;
        }

        /* Ocultado com Slide Up e antecipação suave */
        header.header-hidden {
            transform: translateY(-150%) !important;
            opacity: 0 !important;
            pointer-events: none !important;
            transition: transform 0.55s cubic-bezier(0.5, -0.25, 0.1, 1), opacity 0.4s ease-in !important;
        }
    </style>

    <!-- Regras do Modo Claro (Harpia Light) -->
    <style id="harpia-theme-light">
        html.tema-claro {
            color-scheme: light !important;
        }

        /* 1. Fundo Principal (Farinha de Trigo #FAF5EE) */
        html.tema-claro,
        html.tema-claro body {
            background-color: #FAF5EE !important;
            color: #1A2320 !important;
            min-height: 100% !important;
        }

        /* 2. Cabeçalho Pílula (Pill Shape) no Modo Claro (#E8DDD0) */
        html.tema-claro header {
            background-color: transparent !important;
            border-bottom: none !important;
            color: #1A2320 !important;
        }
        html.tema-claro #header-pill {
            background-color: rgba(232, 221, 208, 0.95) !important;
            border: 1px solid #C8BBA8 !important;
            border-radius: 9999px !important;
            backdrop-filter: blur(12px) !important;
            -webkit-backdrop-filter: blur(12px) !important;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.06) !important;
        }
        html.tema-claro header .text-\[\#F3F6F4\],
        html.tema-claro header span {
            color: #1A2320 !important;
        }
        html.tema-claro header .text-\[\#8C9B9E\] {
            color: #5F6B65 !important;
        }
        html.tema-claro header #status-servidor-badge,
        html.tema-claro header input,
        html.tema-claro header select,
        html.tema-claro header button {
            background-color: #FFFFFF !important;
            border-color: #C8BBA8 !important;
            color: #1A2320 !important;
        }
        html.tema-claro header .bg-\[\#171E26\]\/60 {
            background-color: rgba(255, 255, 255, 0.7) !important;
            border-color: #C8BBA8 !important;
        }

        /* 3. Wrapper e Barra de Filtros Pill Shape no Modo Claro */
        html.tema-claro #barra-filtros-wrapper {
            background-color: transparent !important;
        }
        html.tema-claro #barra-filtros {
            background-color: #E8DDD0 !important;
            border: 1px solid #C8BBA8 !important;
            border-radius: 9999px !important;
            padding: 0.375rem 0.875rem !important;
            width: fit-content !important;
            display: inline-flex !important;
            flex-wrap: wrap !important;
            align-items: center !important;
        }

        /* 4. Cartões, Painéis e Modais (Branco #FFFFFF com Borda #E2D9CC) */
        html.tema-claro .bg-\[\#0A0F0D\],
        html.tema-claro .bg-\[\#0A0F0D\]\/60,
        html.tema-claro .bg-\[\#0A0F0D\]\/95,
        html.tema-claro .bg-\[\#171E26\],
        html.tema-claro .bg-\[\#171E26\]\/20,
        html.tema-claro .bg-\[\#171E26\]\/30,
        html.tema-claro .bg-\[\#171E26\]\/40,
        html.tema-claro .bg-\[\#171E26\]\/50,
        html.tema-claro .bg-\[\#171E26\]\/60,
        html.tema-claro [id^="modal-"] > div,
        html.tema-claro [data-response-panel] {
            --tw-bg-opacity: 1 !important;
            background-color: #FFFFFF !important;
            border-color: #E2D9CC !important;
            color: #1A2320 !important;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
        }

        /* 5. Drawer da Timeline (#FFFFFF) */
        html.tema-claro #timeline-drawer,
        html.tema-claro [id*="timeline"] {
            --tw-bg-opacity: 1 !important;
            background-color: #FFFFFF !important;
            border-color: #E2D9CC !important;
            color: #1A2320 !important;
        }

        /* 6. Rota Ativa na Sidebar & Hover */
        html.tema-claro .ativa-rota {
            --tw-bg-opacity: 1 !important;
            background-color: #E8DDD0 !important;
            color: #C48A00 !important;
            border-left: 3px solid #C48A00 !important;
            font-weight: 700 !important;
        }
        html.tema-claro .hover\:bg-\[\#171E26\]\/30:hover,
        html.tema-claro .hover\:bg-\[\#0A0F0D\]:hover {
            --tw-bg-opacity: 1 !important;
            background-color: #E8DDD0 !important;
        }

        /* 7. Inputs, Textareas, Selects e Pre (#FFFFFF) */
        html.tema-claro input,
        html.tema-claro textarea,
        html.tema-claro select,
        html.tema-claro pre,
        html.tema-claro [id^="modal-"] textarea,
        html.tema-claro [id^="modal-"] pre,
        html.tema-claro [data-response-headers],
        html.tema-claro [data-response-body],
        html.tema-claro [data-snippet-panel] pre {
            --tw-bg-opacity: 1 !important;
            background-color: #FFFFFF !important;
            border-color: #E2D9CC !important;
            color: #1A2320 !important;
        }

        /* Code inline permanece transparente */
        html.tema-claro code {
            background-color: transparent !important;
            border-color: transparent !important;
        }
        html.tema-claro pre code {
            background-color: transparent !important;
        }

        /* 7. Textos Principais e Secundários */
        html.tema-claro .text-\[\#F3F6F4\],
        html.tema-claro .text-white,
        html.tema-claro .text-slate-200,
        html.tema-claro .text-slate-300,
        html.tema-claro h1, html.tema-claro h2, html.tema-claro h3, html.tema-claro h4, html.tema-claro h5, html.tema-claro h6 {
            --tw-text-opacity: 1 !important;
            color: #1A2320 !important;
        }

        html.tema-claro .text-\[\#8C9B9E\],
        html.tema-claro .text-slate-400,
        html.tema-claro .text-slate-500,
        html.tema-claro .text-slate-600 {
            --tw-text-opacity: 1 !important;
            color: #5F6B65 !important;
        }

        /* 8. Botões e Destaques */
        html.tema-claro .text-\[\#F2A900\] {
            --tw-text-opacity: 1 !important;
            color: #C48A00 !important;
        }
        html.tema-claro .text-\[\#D45D34\] {
            --tw-text-opacity: 1 !important;
            color: #B84A2A !important;
        }
        html.tema-claro .text-\[\#00A86B\] {
            --tw-text-opacity: 1 !important;
            color: #1A7A4D !important;
        }
        html.tema-claro .text-blue-400 {
            --tw-text-opacity: 1 !important;
            color: #1A6FB5 !important;
        }
        html.tema-claro .filtro-btn {
            background-color: #FFFFFF !important;
            border-color: #C8BBA8 !important;
            color: #1A2320 !important;
        }
        html.tema-claro .filtro-btn:hover {
            background-color: #FAF5EE !important;
        }
        html.tema-claro .filtro-btn.active-filtro {
            background-color: #C48A00 !important;
            border-color: #C48A00 !important;
            color: #FFFFFF !important;
        }

        html.tema-claro button.bg-\[\#F2A900\],
        html.tema-claro .bg-\[\#F2A900\] {
            --tw-bg-opacity: 1 !important;
            --tw-text-opacity: 1 !important;
            background-color: #C48A00 !important;
            color: #FAF5EE !important;
        }
        html.tema-claro button.bg-\[\#F2A900\]:hover {
            background-color: #A37300 !important;
        }

        html.tema-claro .bg-\[\#D45D34\]\/10 {
            background-color: rgba(184, 74, 42, 0.12) !important;
            border-color: rgba(184, 74, 42, 0.35) !important;
            color: #B84A2A !important;
        }

        html.tema-claro .bg-purple-500\/10,
        html.tema-claro .text-purple-400 {
            background-color: rgba(126, 34, 206, 0.1) !important;
            border-color: rgba(126, 34, 206, 0.3) !important;
            color: #7E22CE !important;
        }

        html.tema-claro .bg-slate-800,
        html.tema-claro .bg-slate-700\/20 {
            --tw-bg-opacity: 1 !important;
            background-color: #FAF5EE !important;
            color: #1A2320 !important;
            border-color: #C8BBA8 !important;
        }

        /* 9. Cabeçalhos de Tabela e Schema */
        html.tema-claro table th,
        html.tema-claro .bg-\[\#0A0F0D\]\/60 {
            --tw-bg-opacity: 1 !important;
            background-color: #FAF5EE !important;
            color: #5F6B65 !important;
            border-color: #C8BBA8 !important;
        }

        /* 10. Bordas Gerais */
        html.tema-claro .border-slate-800,
        html.tema-claro .border-slate-900,
        html.tema-claro .border-slate-900\/30,
        html.tema-claro .border-slate-900\/50,
        html.tema-claro .border-slate-900\/60 {
            --tw-border-opacity: 1 !important;
            border-color: #C8BBA8 !important;
        }

        /* 11. Toasts e Tooltips */
        html.tema-claro #toast-harpia {
            --tw-bg-opacity: 1 !important;
            background-color: #FAF5EE !important;
            border-color: #B84A2A !important;
            color: #1A2320 !important;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.12);
        }
        html.tema-claro #tooltip-float {
            --tw-bg-opacity: 1 !important;
            background-color: #FAF5EE !important;
            border-color: #C48A00 !important;
            color: #1A2320 !important;
        }
        html.tema-claro .relative.group > .absolute {
            --tw-bg-opacity: 1 !important;
            background-color: #FAF5EE !important;
            border-color: #C8BBA8 !important;
            color: #1A2320 !important;
        }
    </style>
</head>
<body class="min-h-full flex flex-col antialiased bg-[#0A0F0D] text-[#F3F6F4] selection:bg-[#F2A900]/30 selection:text-[#F2A900]">
    <!-- Toast de Erro / Notificação -->
    <div id="toast-harpia" class="fixed top-6 left-1/2 -translate-x-1/2 z-[9999] flex items-start gap-3 bg-[#171E26] border border-[#D45D34]/60 rounded-xl px-4 py-3 shadow-2xl font-mono-harpia text-xs text-[#F3F6F4] max-w-sm w-max opacity-0 pointer-events-none transition-all duration-300 transform -translate-y-3">
        <span id="toast-icon" class="text-[#D45D34] text-base shrink-0">⚠️</span>
        <div>
            <div id="toast-titulo" class="font-bold text-[#D45D34] mb-0.5">Erro</div>
            <div id="toast-mensagem" class="text-[#8C9B9E] leading-snug"></div>
        </div>
    </div>
    <div id="banner-servidor-offline" class="hidden bg-[#D45D34] text-white text-xs font-mono-harpia py-1.5 px-4 text-center font-bold shadow-md">
        ⚠️ Servidor caindo ou sem resposta há <span id="segundos-offline">0</span>s. Verifique se o processo Harpia está em execução.
    </div>
    <div id="tooltip-float"></div>
    <script>
    (function(){
        const tip = document.getElementById('tooltip-float');
        document.addEventListener('mouseover', function(e){
            const el = e.target.closest('.has-tip');
            if (!el) return;
            tip.textContent = el.getAttribute('data-tip') || '';
            const r = el.getBoundingClientRect();
            tip.style.left = (r.left + r.width/2) + 'px';

            if (r.top < 100) {
                tip.style.top = (r.bottom + 6) + 'px';
                tip.style.transform = 'translate(-50%, 0%)';
            } else {
                tip.style.top = (r.top - 6) + 'px';
                tip.style.transform = 'translate(-50%, -100%)';
            }
            tip.style.opacity = '1';
        });
        document.addEventListener('mouseout', function(e){
            const el = e.target.closest('.has-tip');
            if (el) tip.style.opacity = '0';
        });
    })();
    </script>

    <!-- Header Pill Shape (Compact Single Line) -->
    <header class="fixed top-0 left-0 right-0 z-40 pt-2.5 px-4 pb-1 header-hidden">
        <div id="header-pill" class="max-w-6xl mx-auto px-5 py-2.5 flex items-center justify-between gap-3 rounded-full border shadow-xl backdrop-blur-md border-slate-800 bg-[#171E26]/95">
            <!-- Brand Logo & Title -->
            <div class="flex items-center gap-2 shrink-0">
                <div id="header-brand-logo" class="flex items-center justify-center w-7 h-7 rounded-lg bg-gradient-to-br from-[#F2A900] to-[#D45D34] shadow-sm shadow-[#F2A900]/20 shrink-0">
                    <span class="text-xs">🦅</span>
                </div>
                <span id="header-brand-title" class="text-xs font-extrabold tracking-tight text-[#F3F6F4]">HARPIA <span class="text-[#F2A900]">CONSOLE</span></span>
            </div>

            <!-- Toolbar (Unificado sem bordas duplas internas, mesmo border-radius do botão Auth) -->
            <div class="flex items-center gap-2 flex-nowrap justify-end shrink-0">
                <!-- Busca (Cmd+K) -->
                <button onclick="abrirSpotlight()" class="has-tip text-[11px] font-medium text-slate-300 hover:text-white px-2.5 py-1 rounded-full hover:bg-[#0A0F0D] transition flex items-center gap-1 font-mono-harpia" data-tip="Busca rápida de rotas (Cmd+K)">
                    <svg class="w-3 h-3 text-[#F2A900]" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
                    <span>Buscar</span>
                    <span class="text-[8px] bg-[#0A0F0D] text-slate-400 px-1 py-0.2 rounded border border-slate-800">⌘K</span>
                </button>

                <!-- Base URL -->
                <div class="flex items-center gap-1 px-1 text-[11px] font-mono-harpia has-tip" data-tip="URL base para requisições">
                    <span class="text-[9px] text-[#8C9B9E]">Base:</span>
                    <input type="text" id="input-base-url" onchange="alterarBaseUrl(this.value)" class="bg-[#0A0F0D] text-[#F2A900] text-[11px] font-bold border border-slate-800/80 rounded-full px-2 py-0.5 focus:outline-none focus:border-[#F2A900]/50" />
                </div>

                <!-- Variáveis {{VAR}} -->
                <button onclick="abrirModalVars()" class="has-tip text-[11px] font-bold text-[#F2A900] hover:bg-[#0A0F0D] px-2 py-1 rounded-full transition font-mono-harpia" data-tip="Variáveis de ambiente {{VAR}}">
                    { }
                </button>

                <div class="h-3.5 w-px bg-slate-800/80"></div>

                <!-- Simulação de Rede -->
                <div class="flex items-center gap-1 px-1">
                    <span class="text-[9px] font-bold text-[#8C9B9E] uppercase tracking-wider font-mono-harpia has-tip" data-tip="Simula latência/erros de rede">Rede</span>
                    <select id="select-correnteza" onchange="mudarCorrenteza(this.value)" class="bg-[#0A0F0D] text-[#F3F6F4] text-[11px] font-mono-harpia border border-slate-800/80 rounded-full px-2 py-0.5 focus:outline-none focus:border-[#F2A900]/50">
                        <option value="bossa">Bossa Nova</option>
                        <option value="3g">3G Interior</option>
                        <option value="instavel">Instável</option>
                    </select>
                </div>

                <!-- Autorização (Auth) -->
                <button onclick="abrirModalAuth()" id="btn-auth" class="has-tip text-[11px] font-medium text-slate-300 hover:text-white px-2.5 py-1 rounded-full hover:bg-[#0A0F0D] transition flex items-center gap-1" data-tip="Configura Bearer Token">
                    <svg class="w-3 h-3 text-[#8C9B9E]" id="icon-auth-lock" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
                    <span id="text-auth">Auth</span>
                </button>

                <div class="h-3.5 w-px bg-slate-800/80"></div>

                <!-- Alternar Tema -->
                <button onclick="alternarTema()" id="btn-tema" class="has-tip p-1 text-[11px] text-slate-400 hover:text-white bg-[#0A0F0D] border border-slate-800/80 rounded-full hover:border-slate-700 transition" data-tip="Alternar Tema Claro / Escuro">
                    ☀️
                </button>

                <!-- Standalone -->
                <button onclick="abrirStandalone()" class="has-tip p-1 text-slate-400 hover:text-white bg-[#0A0F0D] border border-slate-800/80 rounded-full hover:border-slate-700 transition" data-tip="Pop-up Standalone">
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" /></svg>
                </button>

                <!-- Exportar Dropdown -->
                <div class="relative group">
                    <button class="text-[11px] font-semibold text-slate-300 hover:text-[#F2A900] flex items-center gap-1 bg-[#0A0F0D] border border-slate-800/80 px-2.5 py-1 rounded-full hover:border-[#F2A900]/30 transition shadow-sm font-mono-harpia">
                        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
                        <span>Exportar</span>
                        <svg class="w-2.5 h-2.5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
                    </button>
                    <div class="absolute right-0 mt-1 w-44 bg-[#171E26] border border-slate-800 rounded-xl shadow-2xl py-1 hidden group-hover:block z-50 font-mono-harpia text-xs">
                        <a href="/docs.json" target="_blank" class="flex items-center gap-2 px-3 py-2 text-slate-300 hover:bg-[#0A0F0D] hover:text-[#F2A900] transition"><span>📄</span> OpenAPI 3.0</a>
                        <button onclick="exportarPostman()" class="w-full text-left flex items-center gap-2 px-3 py-2 text-slate-300 hover:bg-[#0A0F0D] hover:text-[#F2A900] transition"><span>🚀</span> Postman</button>
                        <button onclick="exportarInsomnia()" class="w-full text-left flex items-center gap-2 px-3 py-2 text-slate-300 hover:bg-[#0A0F0D] hover:text-[#F2A900] transition"><span>☯️</span> Insomnia</button>
                        <button onclick="exportarBruno()" class="w-full text-left flex items-center gap-2 px-3 py-2 text-slate-300 hover:bg-[#0A0F0D] hover:text-[#F2A900] transition"><span>🐶</span> Bruno</button>
                    </div>
                </div>
            </div>
        </div>
    </header>
    <div id="header-spacer" class="h-14 w-full shrink-0 pointer-events-none"></div>

    <!-- Modal de Preview pré-envio -->
    <div id="modal-preview" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center hidden opacity-0 transition-opacity duration-200">
        <div class="bg-[#171E26] border border-slate-800 rounded-xl max-w-2xl w-full p-6 shadow-2xl transform scale-95 transition-transform duration-200">
            <h3 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                <svg class="w-5 h-5 text-[#F2A900]" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                Preview da Requisição
            </h3>
            <p class="text-xs text-[#8C9B9E] leading-relaxed mb-4">Revise a URL, método, cabeçalhos e corpo antes de enviar.</p>
            <div class="space-y-3">
                <div>
                    <div class="text-[10px] text-slate-500 mb-1 font-mono-harpia uppercase">Método + URL</div>
                    <pre id="prev-method-url" class="text-xs font-mono-harpia text-[#F2A900] bg-[#0A0F0D] border border-slate-900 rounded-lg p-2 overflow-x-auto"></pre>
                </div>
                <div>
                    <div class="text-[10px] text-slate-500 mb-1 font-mono-harpia uppercase">Cabeçalhos</div>
                    <pre id="prev-headers" class="text-xs font-mono-harpia text-[#8C9B9E] bg-[#0A0F0D] border border-slate-900 rounded-lg p-2 overflow-x-auto max-h-20"></pre>
                </div>
                <div id="prev-body-wrap" class="hidden">
                    <div class="text-[10px] text-slate-500 mb-1 font-mono-harpia uppercase">Corpo (editável)</div>
                    <textarea id="prev-body" class="w-full text-xs font-mono-harpia text-[#F3F6F4] bg-[#0A0F0D] border border-slate-900 rounded-lg p-2 h-32 focus:outline-none focus:border-[#F2A900]/50"></textarea>
                </div>
            </div>
            <div class="flex gap-2 justify-end mt-4">
                <button onclick="fecharPreview()" class="px-4 py-2 rounded-lg text-xs font-semibold text-[#8C9B9E] hover:text-[#F3F6F4] transition">Cancelar</button>
                <button onclick="confirmarEnvio()" class="px-4 py-2 bg-[#F2A900] text-slate-950 rounded-lg text-xs font-bold hover:bg-[#F2A900]/90 transition shadow-lg">Confirmar e Enviar</button>
            </div>
        </div>
    </div>

    <!-- Modal de Autenticação -->
    <div id="modal-auth" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center hidden opacity-0 transition-opacity duration-200">
        <div class="bg-[#171E26] border border-slate-800 rounded-xl max-w-md w-full p-6 shadow-2xl transform scale-95 transition-transform duration-200">
            <h3 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                <svg class="w-5 h-5 text-[#F2A900]" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" /></svg>
                Configurar Autenticação
            </h3>
            <p class="text-xs text-[#8C9B9E] leading-relaxed mb-4">Adicione um Bearer Token de autorização. Ele será anexado automaticamente a todas as chamadas interativas no cabeçalho <code class="font-mono-harpia text-[#F2A900]">Authorization</code>.</p>
            <div class="space-y-4">
                <div>
                    <label class="block text-xs font-bold text-[#8C9B9E] uppercase tracking-wider mb-2 font-mono-harpia">Bearer Token</label>
                    <input type="text" id="input-token" placeholder="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." class="w-full bg-[#0A0F0D] border border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-[#F2A900]/50 font-mono-harpia">
                </div>
                <div class="flex gap-2 justify-end">
                    <button onclick="fecharModalAuth()" class="px-4 py-2 rounded-lg text-xs font-semibold text-[#8C9B9E] hover:text-[#F3F6F4] transition">Cancelar</button>
                    <button onclick="salvarAuth()" class="px-4 py-2 bg-[#F2A900] text-slate-950 rounded-lg text-xs font-bold hover:bg-[#F2A900]/90 transition shadow-lg">Salvar Credenciais</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Modal de Variáveis de Ambiente -->
    <div id="modal-vars" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center hidden opacity-0 transition-opacity duration-200">
        <div class="bg-[#171E26] border border-slate-800 rounded-xl max-w-lg w-full p-6 shadow-2xl transform scale-95 transition-transform duration-200 font-mono-harpia">
            <h3 class="text-lg font-bold text-white mb-2 flex items-center gap-2">
                <span class="text-[#F2A900] font-bold">{ }</span> Variáveis de Ambiente
            </h3>
            <p class="text-xs text-[#8C9B9E] leading-relaxed mb-4">Defina variáveis como <code class="text-[#F2A900]">{{TOKEN}}</code> para substituição automática.</p>
            <div id="vars-list" class="space-y-2 max-h-60 overflow-y-auto mb-4"></div>
            <div class="flex gap-2 justify-between">
                <button onclick="adicionarLinhaVar()" class="px-3 py-1.5 bg-[#0A0F0D] border border-slate-800 text-[#F2A900] rounded-lg text-xs font-bold hover:bg-slate-900 transition">+ Adicionar Variable</button>
                <div class="flex gap-2">
                    <button onclick="fecharModalVars()" class="px-4 py-2 rounded-lg text-xs font-semibold text-[#8C9B9E] hover:text-[#F3F6F4] transition">Cancelar</button>
                    <button onclick="salvarVars()" class="px-4 py-2 bg-[#F2A900] text-slate-950 rounded-lg text-xs font-bold hover:bg-[#F2A900]/90 transition shadow-lg">Salvar Variáveis</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Modal Spotlight Cmd+K -->
    <div id="modal-spotlight" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-start justify-center pt-24 hidden opacity-0 transition-opacity duration-200">
        <div class="bg-[#171E26] border border-slate-800 rounded-xl max-w-xl w-full p-4 shadow-2xl transform scale-95 transition-transform duration-200 font-mono-harpia">
            <div class="flex items-center gap-3 border-b border-slate-800 pb-3 mb-3">
                <svg class="w-5 h-5 text-[#F2A900]" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
                <input type="text" id="spotlight-input" oninput="filtrarSpotlight(this.value)" placeholder="Buscar rota por nome, método ou caminho..." class="w-full bg-transparent text-sm text-white focus:outline-none" />
                <span class="text-[10px] text-slate-500 bg-[#0A0F0D] px-1.5 py-0.5 rounded border border-slate-800">ESC</span>
            </div>
            <div id="spotlight-results" class="max-h-72 overflow-y-auto space-y-1 scrollbar-thin"></div>
        </div>
    </div>

    <!-- Modal Gerar Teste Harpia -->
    <div id="modal-teste-gerado" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center hidden opacity-0 transition-opacity duration-200">
        <div class="bg-[#171E26] border border-slate-800 rounded-xl max-w-2xl w-full p-6 shadow-2xl transform scale-95 transition-transform duration-200 font-mono-harpia">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-base font-bold text-white flex items-center gap-2">
                    <span class="text-purple-400">🧪</span> Teste Gerado — Harpia Nativo
                </h3>
                <button onclick="fecharModalGerarTeste()" class="text-slate-500 hover:text-white text-xs">✕ Fechar</button>
            </div>
            <p class="text-xs text-[#8C9B9E] mb-4 leading-relaxed">Copie este bloco para seu arquivo <code class="text-[#F2A900]">.hrp</code> de testes.</p>
            <pre id="teste-gerado-code" class="bg-[#0A0F0D] border border-slate-800 rounded-lg p-4 text-xs text-[#F2A900] overflow-x-auto max-h-64 scrollbar-thin whitespace-pre"></pre>
            <div class="flex justify-end mt-4 gap-2">
                <button onclick="fecharModalGerarTeste()" class="px-4 py-2 text-xs text-[#8C9B9E] hover:text-white transition">Cancelar</button>
                <button onclick="copiarTesteGerado()" class="px-4 py-2 bg-purple-500 text-white rounded-lg text-xs font-bold hover:bg-purple-400 transition">📋 Copiar</button>
            </div>
        </div>
    </div>

    <!-- Conteúdo Principal (pt-24 Garante o descolamento do Header Fixed) -->
    <main class="flex-1 max-w-6xl w-full mx-auto px-6 py-10">
        <div id="loading" class="flex flex-col items-center justify-center py-24 gap-4">
            <div class="w-10 h-10 border-2 border-[#F2A900] border-t-transparent rounded-full animate-spin"></div>
            <p class="text-sm text-[#8C9B9E] font-mono-harpia">Verificando rotas e construindo console interativo...</p>
        </div>

        <div id="error" class="hidden max-w-md mx-auto bg-rose-500/5 border border-rose-500/20 rounded-xl p-8 text-center shadow-2xl">
            <div class="w-12 h-12 rounded-full bg-rose-500/10 flex items-center justify-center mx-auto mb-4 border border-rose-500/20 text-[#D45D34]">
                <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            </div>
            <p class="text-base text-[#D45D34] font-bold mb-1">HRP-DOC-5001: Erro de Documentação</p>
            <p class="text-sm text-[#8C9B9E]">Não foi possível conectar com o endpoint /docs.json para buscar a especificação da API.</p>
        </div>

        <div id="content" class="hidden">
            <div class="mb-10 bg-[#171E26]/40 border border-slate-900 rounded-xl p-6 shadow-xl backdrop-blur">
                <h1 id="api-title" class="text-2xl font-extrabold tracking-tight text-white mb-2 font-mono-harpia">API Harpia</h1>
                <p id="api-desc" class="text-[#8C9B9E] text-sm max-w-3xl leading-relaxed"></p>
            </div>

            <!-- Painéis Super Console -->
            <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-8 items-start" id="super-paineis">
                <div class="lg:col-span-1 bg-[#171E26]/30 border border-slate-900 rounded-xl p-5 shadow-lg backdrop-blur hover:scale-[1.01] transition-transform duration-300 flex flex-col justify-between h-full">
                    <div>
                        <h3 class="text-xs font-bold text-[#F2A900] uppercase tracking-wider mb-3 font-mono-harpia flex items-center gap-2">
                            <span>🦅</span> <span class="has-tip" data-tip="Contadores de requisições por rota em tempo real">Voo da Harpia</span>
                        </h3>
                        <p class="text-[10px] text-[#8C9B9E] mb-3 font-mono-harpia">Chamadas em tempo real por rota:</p>
                        <div id="metricas-container" class="space-y-1.5 font-mono-harpia text-xs max-h-[300px] overflow-y-auto scrollbar-thin"></div>
                    </div>
                </div>
                <div class="lg:col-span-1 bg-[#171E26]/30 border border-slate-900 rounded-xl p-5 shadow-lg backdrop-blur hover:scale-[1.01] transition-transform duration-300 flex flex-col justify-between h-full">
                    <div>
                        <h3 class="text-xs font-bold text-[#00A86B] uppercase tracking-wider mb-3 font-mono-harpia flex items-center gap-2">
                            <span>🛡️</span> <span class="has-tip" data-tip="Headers de segurança injetados pelo servidor">Fronteira Segura OWASP</span>
                        </h3>
                        <p class="text-[10px] text-[#8C9B9E] mb-3 font-mono-harpia">Cabeçalhos nativos ativos:</p>
                        <ul id="owasp-container" class="space-y-1.5 font-mono-harpia text-xs text-[#8C9B9E] max-h-[300px] overflow-y-auto scrollbar-thin"></ul>
                    </div>
                </div>
                <div class="lg:col-span-2 md:col-span-3 bg-[#171E26]/30 border border-slate-900 rounded-xl p-5 shadow-lg backdrop-blur hover:scale-[1.01] transition-transform duration-300 flex flex-col justify-between h-full">
                    <div>
                        <h3 class="text-xs font-bold text-[#D45D34] uppercase tracking-wider mb-3 font-mono-harpia flex items-center gap-2">
                            <span>🧠</span> <span class="has-tip" data-tip="Analisa rotas e sugere melhorias REST">Mentor de Design IA</span>
                        </h3>
                        <p class="text-[10px] text-[#8C9B9E] mb-3 font-mono-harpia">Heurísticas detectadas nas rotas:</p>
                        <ul id="mentor-container" class="space-y-1.5 font-mono-harpia text-xs text-[#8C9B9E] max-h-[300px] overflow-y-auto scrollbar-thin"></ul>
                    </div>
                </div>
            </div>

            <!-- Filtros por Método (Sticky no topo com wrapper opaco) -->
            <div id="barra-filtros-wrapper" class="sticky top-0 z-40 pt-3 pb-3 mb-3 transition-colors duration-200">
                <div id="barra-filtros" class="flex flex-wrap items-center gap-2 border border-slate-900 rounded-xl p-3 font-mono-harpia text-xs transition-colors duration-200">
                    <span class="text-[10px] font-bold text-[#8C9B9E] uppercase tracking-wider mr-2">Filtros:</span>
                    <button onclick="filtrarPorMetodo('TODOS')" class="filtro-btn active-filtro px-3 py-1 rounded-lg border border-[#F2A900]/30 bg-[#F2A900]/10 text-[#F2A900] font-bold transition">TODOS</button>
                    <button onclick="filtrarPorMetodo('GET')" class="filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-blue-400 hover:bg-blue-500/10 transition">GET</button>
                    <button onclick="filtrarPorMetodo('POST')" class="filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-[#00A86B] hover:bg-[#00A86B]/10 transition">POST</button>
                    <button onclick="filtrarPorMetodo('PUT')" class="filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-[#F2A900] hover:bg-[#F2A900]/10 transition">PUT</button>
                    <button onclick="filtrarPorMetodo('DELETE')" class="filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-[#D45D34] hover:bg-[#D45D34]/10 transition">DELETE</button>
                    <button onclick="filtrarPorMetodo('PATCH')" class="filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-purple-400 hover:bg-purple-500/10 transition">PATCH</button>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
                <!-- Sidebar Navegação (sticky top-[76px]) -->
                <aside class="md:col-span-1 border-r border-slate-900/60 pr-6 max-h-[calc(100vh-5.5rem)] overflow-y-auto sticky top-[76px] self-start hidden md:block scrollbar-thin">
                    <p class="text-xs font-bold text-[#8C9B9E] uppercase tracking-wider mb-3 font-mono-harpia">Coleção de Rotas</p>
                    <input id="sidebar-search" type="text" placeholder="🔍 Filtrar rota..." oninput="filtrarSidebar(this.value)" class="w-full bg-[#0A0F0D] border border-slate-800 rounded-lg px-2 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-[#F2A900]/50 font-mono-harpia mb-3" />
                    <nav id="sidebar-links" class="space-y-1"></nav>
                    <!-- Histórico -->
                    <div class="mt-6 border-t border-slate-900 pt-4">
                        <div class="flex items-center justify-between mb-2">
                            <p class="text-[10px] font-bold text-[#8C9B9E] uppercase tracking-wider font-mono-harpia">💾 Histórico Salvo</p>
                            <button onclick="limparHistorico()" class="text-[10px] text-slate-600 hover:text-[#D45D34] font-mono-harpia">limpar</button>
                        </div>
                        <div id="historico-container" class="space-y-1"></div>
                    </div>
                </aside>

                <!-- Endpoints Details -->
                <section class="md:col-span-3 space-y-6" id="endpoints-container"></section>
            </div>
        </div>
    </main>

    <!-- Timeline Drawer Bottom -->
    <div id="timeline-drawer" class="fixed bottom-0 left-0 right-0 z-40 bg-[#0A0F0D]/95 border-t border-slate-800 shadow-2xl backdrop-blur font-mono-harpia transition-all duration-300 transform translate-y-36">
        <div class="max-w-6xl mx-auto px-6 py-2 flex items-center justify-between border-b border-slate-900 bg-[#171E26]/50">
            <div class="flex items-center gap-2">
                <span class="text-xs font-bold text-[#F2A900]">⏱️ TIMELINE DA SESSÃO</span>
                <span id="timeline-count" class="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded text-slate-400">0 requisições</span>
            </div>
            <div class="flex items-center gap-2">
                <button onclick="limparTimeline()" class="text-[10px] text-slate-500 hover:text-rose-400">limpar</button>
                <button onclick="alternarTimeline()" id="btn-toggle-timeline" class="text-xs text-[#8C9B9E] hover:text-white px-2 py-0.5 rounded bg-slate-800">▲ Expandir</button>
            </div>
        </div>
        <div id="timeline-list" class="max-w-6xl mx-auto px-6 py-2 max-h-36 overflow-y-auto space-y-1 scrollbar-thin text-xs">
            <p class="text-slate-500 italic text-[11px]">Nenhuma requisição disparada nesta sessão ainda.</p>
        </div>
    </div>

    <script>
        // ===== TOAST DE NOTIFICAÇÃO =====
        let _toastTimer = null;
        function mostrarToast(titulo, mensagem, tipo) {
            const toast = document.getElementById('toast-harpia');
            const tituloEl = document.getElementById('toast-titulo');
            const mensagemEl = document.getElementById('toast-mensagem');
            const iconEl = document.getElementById('toast-icon');
            if (!toast) return;

            if (_toastTimer) clearTimeout(_toastTimer);

            const isClaro = document.documentElement.classList.contains('tema-claro');

            if (tipo === 'sucesso') {
                toast.style.borderColor = isClaro ? '#1A7A4D' : 'rgba(0,168,107,0.5)';
                tituloEl.style.color = isClaro ? '#1A7A4D' : '#00A86B';
                iconEl.textContent = '✅';
            } else if (tipo === 'aviso') {
                toast.style.borderColor = isClaro ? '#E07A2F' : 'rgba(242,169,0,0.5)';
                tituloEl.style.color = isClaro ? '#E07A2F' : '#F2A900';
                iconEl.textContent = '⚠️';
            } else {
                toast.style.borderColor = isClaro ? '#B84A2A' : 'rgba(212,93,52,0.6)';
                tituloEl.style.color = isClaro ? '#B84A2A' : '#D45D34';
                iconEl.textContent = '⚠️';
            }

            tituloEl.textContent = titulo;
            mensagemEl.textContent = mensagem;

            toast.style.opacity = '1';
            toast.style.transform = 'translateX(-50%) translateY(0)';
            toast.style.pointerEvents = 'auto';

            _toastTimer = setTimeout(function() {
                toast.style.opacity = '0';
                toast.style.transform = 'translateX(-50%) translateY(-8px)';
                toast.style.pointerEvents = 'none';
            }, 5000);
        }

        // ===== TRADUTOR DE ERROS JSON =====
        function traduzirErroJSON(rawMsg) {
            const m = rawMsg || '';
            const lcMatch = m.match(/line\s+(\d+)\s+column\s+(\d+)/i);
            const posMatch = m.match(/at position\s+(\d+)/i);
            let onde = '';
            if (lcMatch) {
                onde = ' — linha ' + lcMatch[1] + ', coluna ' + lcMatch[2];
            } else if (posMatch) {
                onde = ' — posição ' + posMatch[1];
            }

            if (/Unexpected end/i.test(m)) return 'JSON incompleto: falta fechar uma chave "}" ou colchete "]"' + onde + '.';
            if (/after property value/i.test(m) || /Expected ','.*after/i.test(m)) return 'Vírgula faltando ou a mais após um valor de propriedade' + onde + '.';
            if (/Unexpected token.*[}\]]/i.test(m)) return 'Fechamento inesperado' + onde + '.';
            if (/Unterminated string/i.test(m)) return 'String não fechada' + onde + '.';
            if (/Expected property name|property name/i.test(m)) return 'Nome de chave inválido' + onde + '.';
            if (/Unexpected token 'n'|Unexpected token n,/i.test(m)) return 'Valor inválido' + onde + '. Use null em minúsculo.';
            if (/Unexpected token 't'|Unexpected token 'f'/i.test(m)) return 'Booleano inválido' + onde + '. Use true ou false em minúsculo.';
            if (/is not valid JSON/i.test(m)) return 'O conteúdo não é um JSON válido' + onde + '.';
            return 'Erro de sintaxe no JSON' + onde + '.';
        }

        let token = localStorage.getItem('harpia_console_token') || '';
        atualizarUIAuth();

        function abrirModalAuth() {
            const modal = document.getElementById('modal-auth');
            modal.classList.remove('hidden');
            setTimeout(() => modal.classList.remove('opacity-0'), 10);
            document.getElementById('input-token').value = token;
        }

        function fecharModalAuth() {
            const modal = document.getElementById('modal-auth');
            modal.classList.add('opacity-0');
            setTimeout(() => modal.classList.add('hidden'), 200);
        }

        function salvarAuth() {
            token = document.getElementById('input-token').value.trim();
            if (token) {
                localStorage.setItem('harpia_console_token', token);
            } else {
                localStorage.removeItem('harpia_console_token');
            }
            atualizarUIAuth();
            fecharModalAuth();
        }

        function atualizarUIAuth() {
            const btn = document.getElementById('btn-auth');
            const icon = document.getElementById('icon-auth-lock');
            const text = document.getElementById('text-auth');
            if (token) {
                btn.className = "has-tip text-[11px] font-bold text-[#00A86B] hover:text-white px-2.5 py-1 rounded-full hover:bg-[#0A0F0D] transition flex items-center gap-1";
                icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />';
                text.textContent = "Autorizado";
            } else {
                btn.className = "has-tip text-[11px] font-medium text-slate-300 hover:text-white px-2.5 py-1 rounded-full hover:bg-[#0A0F0D] transition flex items-center gap-1";
                icon.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />';
                text.textContent = "Autorizar";
            }
        }

        // ===== PREVIEW PRÉ-ENVIO =====
        let _previewMethod = '', _previewUrl = '', _previewBody = null, _previewBtn = null;
        function abrirPreview(method, pattern, btn) {
            const card = btn.closest('.scroll-mt-20');
            let url = pattern;
            card.querySelectorAll('[data-route-param]').forEach(input => {
                const p = input.getAttribute('data-route-param');
                const val = input.value.trim() || ':' + p;
                url = url.replace(':' + p, encodeURIComponent(val));
            });
            _previewMethod = method;
            _previewUrl = url;
            _previewBtn = btn;
            const bodyArea = card.querySelector('[data-request-body]');
            _previewBody = bodyArea ? bodyArea.value.trim() : '';

            document.getElementById('prev-method-url').textContent = method.toUpperCase() + '  ' + url;
            let h = 'Content-Type: application/json';
            if (token) h += '\nAuthorization: Bearer ' + token.slice(0, 20) + (token.length > 20 ? '...' : '');
            document.getElementById('prev-headers').textContent = h;
            const bodyWrap = document.getElementById('prev-body-wrap');
            if (_previewBody) {
                bodyWrap.classList.remove('hidden');
                document.getElementById('prev-body').value = _previewBody;
            } else {
                bodyWrap.classList.add('hidden');
            }
            const modal = document.getElementById('modal-preview');
            modal.classList.remove('hidden');
            setTimeout(() => modal.classList.remove('opacity-0'), 10);
        }
        function fecharPreview() {
            const modal = document.getElementById('modal-preview');
            modal.classList.add('opacity-0');
            setTimeout(() => modal.classList.add('hidden'), 200);
        }
        function confirmarEnvio() {
            const bodyArea = _previewBtn.closest('.scroll-mt-20').querySelector('[data-request-body]');
            const prevBody = document.getElementById('prev-body');
            if (bodyArea && !document.getElementById('prev-body-wrap').classList.contains('hidden')) {
                const edited = prevBody.value.trim();
                if (edited) {
                    try { JSON.parse(edited); } catch(e) {
                        mostrarToast('JSON Inválido', traduzirErroJSON(e.message));
                        return;
                    }
                }
                bodyArea.value = edited;
            }
            fecharPreview();
            enviarTeste(_previewMethod, _previewUrl, _previewBtn);
        }

        // ===== MODO TEATRO =====
        function teatroFeedback(btn, msg, cor) {
            const original = btn.innerHTML;
            btn.innerHTML = msg;
            btn.style.background = cor;
            btn.style.color = '#F3F6F4';
            setTimeout(() => {
                btn.innerHTML = original;
                btn.style.background = '';
                btn.style.color = '';
            }, 1800);
        }

        let _openapiData = null;

        function preencherTeatro(btn) {
            var card = btn.closest('[id]');
            var bodyArea = card.querySelector('[data-request-body]');
            var inputs = card.querySelectorAll('[data-route-param]');
            var path = card.dataset.path || '';
            var method = (card.dataset.method || 'post').toLowerCase();

            if (!bodyArea && inputs.length === 0) {
                teatroFeedback(btn, '🎭 sem campos', '#8C9B9E');
                return;
            }

            var routeInfo = (_openapiData && _openapiData.paths && _openapiData.paths[path]) ? _openapiData.paths[path][method] : null;

            if (bodyArea) {
                if (routeInfo && routeInfo.requestBody && routeInfo.requestBody.content && routeInfo.requestBody.content['application/json']) {
                    var contentObj = routeInfo.requestBody.content['application/json'];
                    var payload = contentObj.example || contentObj.schema || {};
                    bodyArea.value = JSON.stringify(payload, null, 2);
                } else {
                    bodyArea.value = "{}";
                }
            }

            inputs.forEach(function(input){
                input.value = "100";
            });

            teatroFeedback(btn, '✓ preenchido', '#00A86B');
        }

        fetch('/docs.json')
            .then(res => {
                if (!res.ok) throw new Error();
                return res.json();
            })
            .then(data => {
                _openapiData = data;
                document.getElementById('loading').classList.add('hidden');
                document.getElementById('content').classList.remove('hidden');

                document.getElementById('api-title').textContent = data.info.title || 'Harpia API';
                document.getElementById('api-desc').textContent = data.info.description || 'Nenhuma descrição fornecida.';

                const container = document.getElementById('endpoints-container');
                const sidebar = document.getElementById('sidebar-links');

                if (!data.paths || Object.keys(data.paths).length === 0) {
                    container.innerHTML =
                        '<div class="text-center py-16 text-[#8C9B9E] italic bg-[#171E26]/20 border border-slate-900 rounded-xl shadow-inner font-mono-harpia">' +
                            '[!] Nenhuma rota registrada neste servidor HTTP.' +
                        '</div>';
                    return;
                }

                for (const [path, methods] of Object.entries(data.paths)) {
                    for (const [method, info] of Object.entries(methods)) {
                        const mLower = method.toLowerCase();
                        const id = mLower + '-' + path.replace(/[^a-zA-Z0-9]/g, '-');
                        const params = (path.match(/:[a-zA-Z0-9_]+/g) || []).map(p => p.slice(1));
                        const temSchema = !!(info.requestBody && info.requestBody.content && info.requestBody.content['application/json']);

                        const badges = {
                            get: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
                            post: 'bg-[#00A86B]/10 text-[#00A86B] border-[#00A86B]/20',
                            delete: 'bg-[#D45D34]/10 text-[#D45D34] border-[#D45D34]/20',
                            put: 'bg-[#F2A900]/10 text-[#F2A900] border-[#F2A900]/20',
                            patch: 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                        };
                        const badgeColor = badges[mLower] || 'bg-slate-500/10 text-slate-400 border-slate-500/20';

                        const schemaBadge = temSchema
                            ? '<span class="ml-auto text-[9px] font-bold px-1.5 py-0.5 rounded border bg-[#F2A900]/10 text-[#F2A900] border-[#F2A900]/20 font-mono-harpia has-tip" data-tip="Metadados declarados: schema, exemplo e respostas disponíveis">📋 Schema</span>'
                            : '<span class="ml-auto text-[9px] font-bold px-1.5 py-0.5 rounded border bg-slate-700/20 text-slate-600 border-slate-800 font-mono-harpia has-tip" data-tip="Sem metadados declarados.">⬜ Manual</span>';

                        const isDeprecated = !!info.deprecated;
                        const sideLink = document.createElement('a');
                        sideLink.href = '#' + id;
                        sideLink.dataset.sidebarPath = (method + ' ' + path).toLowerCase();
                        sideLink.className = 'sidebar-rota group flex items-center gap-2 py-2 px-3 text-xs text-[#8C9B9E] hover:text-[#F3F6F4] rounded-lg hover:bg-[#171E26]/40 border border-transparent hover:border-slate-800/50 transition font-medium font-mono-harpia' + (isDeprecated ? ' opacity-60' : '');
                        sideLink.innerHTML =
                            (isDeprecated ? '<span class="text-[9px] text-[#F2A900]">⚠</span>' : '') +
                            '<span class="text-[10px] font-bold uppercase tracking-wider w-10 text-slate-500 group-hover:text-[#F2A900]">' + method + '</span>' +
                            '<span class="truncate' + (isDeprecated ? ' line-through opacity-60' : '') + '">' + path + '</span>';
                        sidebar.appendChild(sideLink);

                        let schemaVisualHTML = '';
                        if (temSchema) {
                            const cObj = info.requestBody.content['application/json'];
                            const props = (cObj.schema && cObj.schema.properties) ? cObj.schema.properties : {};
                            const required = (cObj.schema && cObj.schema.required) || [];
                            const propKeys = Object.keys(props);
                            if (propKeys.length > 0) {
                                schemaVisualHTML = '<div class="mb-4 border border-slate-800/50 rounded-lg overflow-hidden">' +
                                    '<div class="bg-[#0A0F0D]/60 px-3 py-2 border-b border-slate-800/50 text-[10px] font-bold text-[#8C9B9E] uppercase tracking-wider font-mono-harpia">Schema do Corpo</div>' +
                                    '<table class="w-full text-xs font-mono-harpia">' +
                                    '<thead><tr class="border-b border-slate-800/30">' +
                                        '<th class="text-left px-3 py-2 text-[10px] text-slate-600">Campo</th>' +
                                        '<th class="text-left px-3 py-2 text-[10px] text-slate-600">Tipo</th>' +
                                        '<th class="text-left px-3 py-2 text-[10px] text-slate-600">Obrigatório</th>' +
                                    '</tr></thead><tbody>';
                                propKeys.forEach(k => {
                                    const t = (props[k] && props[k].type) || 'any';
                                    const req = required.includes(k);
                                    schemaVisualHTML += '<tr class="border-b border-slate-900/30 hover:bg-[#171E26]/20">' +
                                        '<td class="px-3 py-1.5 text-[#F2A900]">' + k + '</td>' +
                                        '<td class="px-3 py-1.5 text-blue-400">' + t + '</td>' +
                                        '<td class="px-3 py-1.5">' + (req ? '<span class="text-[#00A86B]">✓</span>' : '<span class="text-slate-600">—</span>') + '</td>' +
                                    '</tr>';
                                });
                                schemaVisualHTML += '</tbody></table></div>';
                            }
                        }

                        let paramsHTML = '';
                        if (params.length > 0) {
                            paramsHTML += '<div class="mb-4"><h4 class="text-xs font-bold text-[#8C9B9E] uppercase tracking-wider mb-2 font-mono-harpia">Parâmetros de Rota</h4><div class="grid grid-cols-1 md:grid-cols-2 gap-3">';
                            params.forEach(p => {
                                paramsHTML +=
                                    '<div>' +
                                        '<label class="block text-[10px] font-mono-harpia text-[#8C9B9E] mb-1">:' + p + '</label>' +
                                        '<input type="text" data-route-param="' + p + '" placeholder="ex: 100" class="w-full bg-[#0A0F0D] border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-[#F2A900]/50 font-mono-harpia">' +
                                    '</div>';
                            });
                            paramsHTML += '</div></div>';
                        }

                        let bodyHTML = '';
                        if (['post', 'put', 'patch', 'delete'].includes(mLower)) {
                            let placeholderEx = '{}';
                            if (temSchema) {
                                let cObj = info.requestBody.content['application/json'];
                                placeholderEx = JSON.stringify(cObj.example || cObj.schema || {}, null, 2);
                            }
                            const semSchemaAviso = !temSchema ? '<div class="mb-2 flex items-center gap-2 text-[10px] font-mono-harpia text-slate-600"><span>⬜</span><span>Sem schema declarado — passe um 3º argumento em <code class="text-[#F2A900]">srv.postar()</code>.</span></div>' : '';
                            bodyHTML =
                                '<div class="mb-4">' +
                                    semSchemaAviso +
                                    '<h4 class="text-xs font-bold text-[#8C9B9E] uppercase tracking-wider mb-2 font-mono-harpia">Corpo da Requisição (JSON)</h4>' +
                                    '<textarea data-request-body class="w-full bg-[#0A0F0D] border border-slate-800 rounded-lg p-3 text-xs text-slate-200 focus:outline-none focus:border-[#F2A900]/50 font-mono-harpia h-24" placeholder="' + placeholderEx.replace(/"/g, '&quot;') + '"></textarea>' +
                                '</div>';
                        }

                        const snippetPanelHTML =
                            '<div data-snippet-panel class="hidden mb-4 bg-[#0A0F0D] border border-slate-800 rounded-lg p-3">' +
                                '<div class="flex items-center justify-between mb-2">' +
                                    '<div class="flex gap-1 font-mono-harpia text-[10px]">' +
                                        '<button onclick="mudarLang(this, \'' + id + '\', \'curl\')" class="snippet-lang active-lang px-2 py-0.5 rounded bg-[#F2A900]/20 text-[#F2A900] font-bold border border-[#F2A900]/20">cURL</button>' +
                                        '<button onclick="mudarLang(this, \'' + id + '\', \'js\')" class="snippet-lang px-2 py-0.5 rounded text-[#8C9B9E] hover:text-white">JavaScript</button>' +
                                        '<button onclick="mudarLang(this, \'' + id + '\', \'python\')" class="snippet-lang px-2 py-0.5 rounded text-[#8C9B9E] hover:text-white">Python</button>' +
                                        '<button onclick="mudarLang(this, \'' + id + '\', \'harpia\')" class="snippet-lang px-2 py-0.5 rounded text-[#8C9B9E] hover:text-white">Harpia</button>' +
                                    '</div>' +
                                    '<button onclick="copiarSnippet(this)" class="text-[10px] text-[#8C9B9E] hover:text-[#F2A900] font-mono-harpia">📋 Copiar</button>' +
                                '</div>' +
                                '<pre data-snippet-code class="text-xs font-mono-harpia text-[#F2A900] overflow-x-auto whitespace-pre-wrap scrollbar-thin max-h-48"></pre>' +
                            '</div>';

                        const el = document.createElement('div');
                        el.id = id;
                        el.dataset.path = path;
                        el.dataset.method = method;
                        el.className = 'scroll-mt-20 bg-[#171E26]/30 border ' + (isDeprecated ? 'border-[#F2A900]/20' : 'border-slate-900') + ' rounded-xl p-6 hover:border-slate-800/80 transition shadow-lg backdrop-blur relative overflow-hidden';
                        const latencyBadge = '<span data-latency-badge class="hidden ml-2 text-[10px] font-mono-harpia font-bold px-2 py-0.5 rounded border bg-[#00A86B]/10 text-[#00A86B] border-[#00A86B]/20 transition-all"></span>';

                        const deprecatedBanner = isDeprecated
                            ? '<div class="absolute top-0 right-0 left-0 bg-[#F2A900]/10 border-b border-[#F2A900]/30 px-4 py-1.5 flex items-center gap-2 text-[11px] font-mono-harpia font-bold text-[#F2A900]">'
                              + '<span class="text-[#F2A900]">⚠️ DEPRECATED</span>'
                              + (info['x-harpia-deprecacao'] ? ' <span class="font-normal text-[#8C9B9E]">— ' + info['x-harpia-deprecacao'] + '</span>' : '')
                              + '</div>'
                            : '';
                        const deprecatedPaddingTop = isDeprecated ? ' pt-10' : '';

                        el.innerHTML =
                            deprecatedBanner +
                            '<div class="flex items-center gap-3 border-b border-slate-900 pb-4 mb-4' + deprecatedPaddingTop + '">' +
                                '<span class="px-2.5 py-1 text-xs font-bold uppercase tracking-wider rounded border ' + badgeColor + ' font-mono-harpia">' + method + '</span>' +
                                '<code class="text-base font-semibold font-mono-harpia ' + (isDeprecated ? 'line-through opacity-60 text-[#8C9B9E]' : 'text-[#F2A900]') + '">' + path + '</code>' +
                                latencyBadge +
                                schemaBadge +
                            '</div>' +
                            '<h3 class="text-lg font-bold text-[#F3F6F4] mb-2">' + (info.summary || ('Endpoint ' + method.toUpperCase() + ' ' + path)) + '</h3>' +
                            '<p class="text-sm text-[#8C9B9E] leading-relaxed mb-4">' + (info.description || 'Nenhuma descrição detalhada disponível para este endpoint.') + '</p>' +

                            schemaVisualHTML +

                            '<div class="border-t border-slate-900/60 pt-4 mt-4">' +
                                '<div class="flex items-center justify-between mb-4">' +
                                    '<div class="flex items-center gap-3">' +
                                        '<button onclick="alternarTabDoc(this,\'teste\')" class="doc-tab-btn text-xs font-bold text-[#F2A900] uppercase tracking-wider font-mono-harpia border-b-2 border-[#F2A900] pb-1">Console de Teste</button>' +
                                        '<button onclick="alternarTabDoc(this,\'snippet\')" class="doc-tab-btn text-xs font-bold text-[#8C9B9E] uppercase tracking-wider font-mono-harpia border-b-2 border-transparent pb-1 hover:text-[#F3F6F4]">Snippet de Código</button>' +
                                    '</div>' +
                                    '<div class="flex items-center gap-2">' +
                                        '<button onclick="salvarNoHistorico(this)" class="has-tip px-3 py-1.5 bg-slate-700/20 border border-slate-800 text-slate-400 text-xs font-bold rounded-lg hover:bg-slate-700/40 transition font-mono-harpia" data-tip="Salva esta requisição no histórico local">💾</button>' +
                                        '<button onclick="preencherTeatro(this)" class="has-tip px-3 py-1.5 bg-[#D45D34]/10 border border-[#D45D34]/30 text-[#D45D34] text-xs font-bold rounded-lg hover:bg-[#D45D34]/20 transition font-mono-harpia flex items-center gap-1" data-tip="Preenche campos com dados de teste">' +
                                            '🎭 Teatro' +
                                        '</button>' +
                                        '<button onclick="abrirModalGerarTeste(this)" class="has-tip px-3 py-1.5 bg-purple-500/10 border border-purple-500/30 text-purple-400 text-xs font-bold rounded-lg hover:bg-purple-500/20 transition font-mono-harpia" data-tip="Gera bloco testar{} nativo">🧪 Gerar Teste</button>' +
                                        '<button onclick="abrirPreview(\'' + method + '\', \'' + path + '\', this)" class="px-4 py-1.5 bg-[#F2A900] text-slate-950 font-bold text-xs rounded-lg hover:bg-[#F2A900]/90 transition shadow-md flex items-center gap-1.5">' +
                                            '<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>' +
                                            'Enviar' +
                                        '</button>' +
                                    '</div>' +
                                '</div>' +

                                snippetPanelHTML +

                                '<div data-teste-panel>' +

                                paramsHTML +
                                bodyHTML +

                                '<div data-response-panel class="hidden border border-slate-900 bg-[#0A0F0D]/60 rounded-xl p-4 shadow-inner mt-4 font-mono-harpia text-xs">' +
                                    '<div class="flex items-center justify-between border-b border-slate-900 pb-2 mb-2">' +
                                        '<div class="flex items-center gap-2">' +
                                            '<span class="font-bold text-slate-400">STATUS DA RESPOSTA</span>' +
                                            '<button onclick="fixarRespostaA(this)" class="px-2 py-0.5 bg-slate-800 text-[#F2A900] rounded text-[10px] hover:bg-slate-700 transition">📌 Fixar como A</button>' +
                                        '</div>' +
                                        '<div class="flex items-center gap-3">' +
                                            '<span data-response-time class="text-[10px] text-[#8C9B9E]"></span>' +
                                            '<span data-response-status class="font-bold uppercase px-2 py-0.5 rounded"></span>' +
                                        '</div>' +
                                    '</div>' +
                                    '<div class="space-y-3">' +
                                        '<div>' +
                                            '<div class="text-[10px] text-slate-500 mb-1">CABEÇALHOS</div>' +
                                            '<pre data-response-headers class="text-[#8C9B9E] overflow-x-auto bg-[#171E26]/20 p-2 rounded border border-slate-900 max-h-24 scrollbar-thin"></pre>' +
                                        '</div>' +
                                        '<div>' +
                                            '<div class="flex items-center justify-between mb-1">' +
                                                '<div class="text-[10px] text-slate-500">CORPO DA RESPOSTA</div>' +
                                                '<span data-[#diff-badge] class="hidden text-[10px] text-[#F2A900] font-bold">⚡ Modo Diff Ativo</span>' +
                                            '</div>' +
                                            '<div data-diff-container class="hidden grid grid-cols-2 gap-2 mb-2 text-xs font-mono-harpia">' +
                                                '<div><div class="text-[9px] text-[#F2A900] mb-1 font-bold">RESPOSTA A (FIXADA)</div><pre data-diff-a class="bg-[#171E26]/40 p-2 rounded border border-slate-800 max-h-48 overflow-y-auto whitespace-pre-wrap text-slate-300"></pre></div>' +
                                                '<div><div class="text-[9px] text-[#00A86B] mb-1 font-bold">RESPOSTA B (ATUAL)</div><pre data-diff-b class="bg-[#171E26]/40 p-2 rounded border border-slate-800 max-h-48 overflow-y-auto whitespace-pre-wrap text-slate-300"></pre></div>' +
                                            '</div>' +
                                            '<pre data-response-body class="text-[#F3F6F4] overflow-x-auto bg-[#171E26]/20 p-3 rounded border border-slate-900 max-h-48 scrollbar-thin whitespace-pre-wrap"></pre>' +
                                        '</div>' +
                                    '</div>' +
                                '</div>' +
                                '</div>' +
                            '</div>';
                        container.appendChild(el);

                        setTimeout(() => {
                            const snipEl = el.querySelector('[data-snippet-code]');
                            if (snipEl) {
                                const code = gerarSnippet('curl', method, path, info, el);
                                snipEl.innerHTML = destacarSintaxe('curl', code);
                            }
                        }, 0);
                    }
                }
            })
            .catch(() => {
                document.getElementById('loading').classList.add('hidden');
                document.getElementById('error').classList.remove('hidden');
            });

        function enviarTeste(method, url, btn) {
            const card = btn.closest('.scroll-mt-20');
            let bodyText = '';
            const bodyArea = card.querySelector('[data-request-body]');
            if (bodyArea) {
                bodyText = bodyArea.value.trim();
                bodyText = interpolarVars(bodyText);
                if (bodyText) {
                    try { JSON.parse(bodyText); } catch(e) {
                        mostrarToast('JSON Inválido', traduzirErroJSON(e.message));
                        return;
                    }
                }
            }
            const headers = { 'Content-Type': 'application/json' };
            if (token) headers['Authorization'] = 'Bearer ' + interpolarVars(token);
            _executarEnvio(method, url, btn, card, bodyText, headers);
        }

        function _executarEnvio(method, url, btn, card, bodyText, headers) {
            const responsePanel = card.querySelector('[data-response-panel]');
            const rTime = card.querySelector('[data-response-time]');
            const rStatus = card.querySelector('[data-response-status]');
            const rHeaders = card.querySelector('[data-response-headers]');
            const rBody = card.querySelector('[data-response-body]');

            btn.disabled = true;
            btn.classList.add('opacity-50');

            const finalBody = interpolarVars(bodyText);
            const bodyOption = (['POST','PUT','PATCH','DELETE'].includes(method.toUpperCase()) && finalBody) ? finalBody : undefined;

            const start = performance.now();
            aplicarCorrenteza().then(() => {
                return fetch(url, {
                    method: method.toUpperCase(),
                    headers: headers,
                    body: bodyOption
                });
            })
            .then(async res => {
                const duration = (performance.now() - start).toFixed(1);
                responsePanel.classList.remove('hidden');

                rStatus.textContent = res.status + ' ' + res.statusText;
                if (res.ok) {
                    rStatus.className = 'font-bold uppercase px-2 py-0.5 rounded bg-[#00A86B]/10 text-[#00A86B] border border-[#00A86B]/20';
                } else {
                    rStatus.className = 'font-bold uppercase px-2 py-0.5 rounded bg-[#D45D34]/10 text-[#D45D34] border border-[#D45D34]/20';
                }
                rTime.textContent = duration + ' ms';

                const latBadge = card.querySelector('[data-latency-badge]');
                if (latBadge) {
                    latBadge.textContent = '⚡ ' + duration + ' ms · ' + res.status;
                    latBadge.classList.remove('hidden');
                    if (res.ok) {
                        latBadge.className = 'ml-2 text-[10px] font-mono-harpia font-bold px-2 py-0.5 rounded border bg-[#00A86B]/10 text-[#00A86B] border-[#00A86B]/20';
                    } else {
                        latBadge.className = 'ml-2 text-[10px] font-mono-harpia font-bold px-2 py-0.5 rounded border bg-[#D45D34]/10 text-[#D45D34] border-[#D45D34]/20';
                    }
                }

                let headersText = '';
                for (let [key, val] of res.headers.entries()) {
                    headersText += key + ': ' + val + '\n';
                }
                rHeaders.textContent = headersText || '(Nenhum)';

                const text = await res.text();
                let bodyFormatado = text;
                try {
                    const parsed = JSON.parse(text);
                    bodyFormatado = JSON.stringify(parsed, null, 2);
                } catch(e) {
                    bodyFormatado = text || '(Corpo Vazio)';
                }
                rBody.textContent = bodyFormatado;

                const diffContainer = card.querySelector('[data-diff-container]');
                const diffBPre = card.querySelector('[data-diff-b]');
                if (diffContainer && !diffContainer.classList.contains('hidden') && diffBPre) {
                    diffBPre.textContent = bodyFormatado;
                }

                registrarTimeline(method, url, res.status, duration, bodyFormatado);
            })
            .catch(err => {
                responsePanel.classList.remove('hidden');
                rStatus.textContent = 'Erro de Conexão';
                rStatus.className = 'font-bold uppercase px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20';
                rTime.textContent = '0 ms';
                rHeaders.textContent = '(Nenhum)';
                rBody.textContent = 'HRP-NET-5003: Falha ao enviar requisição. Verifique sua conexão com o servidor.';
                const latBadge = card.querySelector('[data-latency-badge]');
                if (latBadge) {
                    latBadge.textContent = '⚡ Erro';
                    latBadge.className = 'ml-2 text-[10px] font-mono-harpia font-bold px-2 py-0.5 rounded border bg-rose-500/10 text-rose-400 border-rose-500/20';
                    latBadge.classList.remove('hidden');
                }
            })
            .finally(() => {
                btn.disabled = false;
                btn.classList.remove('opacity-50');
            });
        }

        // ===== TEMA CLARO / ESCURO =====
        (function() {
            const salvo = localStorage.getItem('harpia_tema');
            if (salvo === 'claro') {
                document.documentElement.classList.remove('tema-escuro');
                document.documentElement.classList.add('tema-claro');
                const btn = document.getElementById('btn-tema');
                if (btn) btn.textContent = '🌙';
            }
        })();

        function alternarTema() {
            const isEscuro = document.documentElement.classList.contains('tema-escuro');
            const btn = document.getElementById('btn-tema');
            if (isEscuro) {
                document.documentElement.classList.remove('tema-escuro');
                document.documentElement.classList.add('tema-claro');
                localStorage.setItem('harpia_tema', 'claro');
                if (btn) btn.textContent = '🌙';
            } else {
                document.documentElement.classList.remove('tema-claro');
                document.documentElement.classList.add('tema-escuro');
                localStorage.setItem('harpia_tema', 'escuro');
                if (btn) btn.textContent = '☀️';
            }
        }

        function fixarRespostaA(btn) {
            const card = btn.closest('[data-response-panel]');
            if (!card) return;
            const rBody = card.querySelector('[data-response-body]');
            const diffContainer = card.querySelector('[data-diff-container]');
            const diffAPre = card.querySelector('[data-diff-a]');
            if (!rBody || !diffContainer || !diffAPre) return;

            const textoAtual = rBody.textContent;
            if (!textoAtual || textoAtual === '(Corpo Vazio)') return;

            diffAPre.textContent = textoAtual;
            diffContainer.classList.remove('hidden');
            diffContainer.style.display = 'grid';
            btn.textContent = '📌 A fixada — dispare outra chamada';
            btn.classList.add('opacity-50');
            btn.disabled = true;
        }

        // ===== TIMELINE =====
        let _timelineEntries = [];
        let _timelineAberta = false;

        function registrarTimeline(method, url, status, duration, body) {
            const hora = new Date().toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
            _timelineEntries.unshift({ method, url, status, duration, body, hora });

            const countEl = document.getElementById('timeline-count');
            if (countEl) countEl.textContent = _timelineEntries.length + ' requisição' + (_timelineEntries.length > 1 ? 'ões' : '');

            const drawer = document.getElementById('timeline-drawer');
            if (drawer) {
                drawer.style.transform = 'translateY(0)';
            }

            renderizarTimeline();
        }

        function renderizarTimeline() {
            const lista = document.getElementById('timeline-list');
            if (!lista) return;
            if (_timelineEntries.length === 0) {
                lista.innerHTML = '<p class="text-slate-500 italic text-[11px]">Nenhuma requisição disparada nesta sessão ainda.</p>';
                return;
            }
            const methodColors = { 'GET': 'text-blue-400', 'POST': 'text-[#00A86B]', 'PUT': 'text-[#F2A900]', 'PATCH': 'text-purple-400', 'DELETE': 'text-[#D45D34]' };
            lista.innerHTML = _timelineEntries.map(function(e, i) {
                const cor = e.status >= 200 && e.status < 300 ? 'text-[#00A86B]' : 'text-[#D45D34]';
                const metodoCor = methodColors[e.method.toUpperCase()] || 'text-slate-400';
                return '<div class="flex items-center gap-3 py-1 border-b border-slate-900/50 text-[11px] hover:bg-[#171E26]/30 rounded px-1 cursor-pointer" title="' + e.body.substring(0, 200) + '">' +
                    '<span class="text-slate-600">' + e.hora + '</span>' +
                    '<span class="font-bold ' + metodoCor + '">' + e.method.toUpperCase() + '</span>' +
                    '<span class="text-slate-400 truncate max-w-xs">' + e.url + '</span>' +
                    '<span class="ml-auto font-bold ' + cor + '">' + e.status + '</span>' +
                    '<span class="text-slate-500">' + e.duration + ' ms</span>' +
                '</div>';
            }).join('');
        }

        function alternarTimeline() {
            const drawer = document.getElementById('timeline-drawer');
            const btn = document.getElementById('btn-toggle-timeline');
            if (!drawer) return;
            _timelineAberta = !_timelineAberta;
            if (_timelineAberta) {
                drawer.style.transform = 'translateY(0)';
                if (btn) btn.textContent = '▼ Recolher';
            } else {
                drawer.style.transform = 'translateY(144px)';
                if (btn) btn.textContent = '▲ Expandir';
            }
        }

        function limparTimeline() {
            _timelineEntries = [];
            const countEl = document.getElementById('timeline-count');
            if (countEl) countEl.textContent = '0 requisições';
            renderizarTimeline();
        }

        // ===== GERAR TESTE =====
        function abrirModalGerarTeste(btn) {
            const card = btn.closest('.scroll-mt-20');
            if (!card) return;
            const method = card.dataset.method || 'GET';
            const path = card.dataset.path || '/';
            const bodyArea = card.querySelector('[data-request-body]');
            const bodyText = bodyArea ? bodyArea.value.trim() : '';

            const info = (_openapiData && _openapiData.paths && _openapiData.paths[path])
                ? (_openapiData.paths[path][method.toLowerCase()] || {})
                : {};

            const codigo = gerarTesteHarpia(method, path, info, bodyText);
            const pre = document.getElementById('teste-gerado-code');
            if (pre) pre.textContent = codigo;

            const modal = document.getElementById('modal-teste-gerado');
            modal.classList.remove('hidden');
            setTimeout(function() {
                modal.classList.remove('opacity-0');
                modal.querySelector('div').classList.remove('scale-95');
            }, 10);
        }

        function fecharModalGerarTeste() {
            const modal = document.getElementById('modal-teste-gerado');
            modal.classList.add('opacity-0');
            modal.querySelector('div').classList.add('scale-95');
            setTimeout(function() { modal.classList.add('hidden'); }, 200);
        }

        function gerarTesteHarpia(method, path, info, bodyText) {
            const baseUrl = (typeof _baseUrl !== 'undefined' ? _baseUrl : ('http://localhost:' + location.port)) || 'http://localhost:3000';
            const url = baseUrl + path;
            const summary = info.summary || (method.toUpperCase() + ' ' + path);

            let corpo = bodyText;
            if (!corpo && info.requestBody) {
                try {
                    const cObj = info.requestBody.content['application/json'];
                    corpo = JSON.stringify(cObj.example || cObj.schema || {}, null, 4);
                } catch(e) {
                    corpo = '{}';
                }
            }

            const temCorpo = ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method.toUpperCase()) && corpo;
            const linhasCorpo = temCorpo
                ? '\n    var corpo = ' + (corpo || '{}') + '\n    var resp = requisitar("' + method.toUpperCase() + '", "' + url + '", corpo)'
                : '\n    var resp = requisitar("' + method.toUpperCase() + '", "' + url + '")';

            return 'testar "' + summary + ' retorna 200" {' +
                linhasCorpo + '\n' +
                '    assegura(resp.status == 200)\n' +
                '    assegura(resp.corpo != "")\n' +
                '}';
        }

        function copiarTesteGerado() {
            const pre = document.getElementById('teste-gerado-code');
            if (!pre) return;
            navigator.clipboard.writeText(pre.textContent).then(function() {
                const btn = document.querySelector('#modal-teste-gerado button[onclick="copiarTesteGerado()"]');
                if (btn) { const orig = btn.textContent; btn.textContent = '✓ Copiado!'; setTimeout(function(){ btn.textContent = orig; }, 1500); }
            });
        }

        function filtrarSidebar(q) {
            const rotas = document.querySelectorAll('.sidebar-rota');
            rotas.forEach(a => {
                const match = a.dataset.sidebarPath.includes(q.toLowerCase());
                a.style.display = match ? '' : 'none';
            });
        }

        function alternarTabDoc(btn, tab) {
            const card = btn.closest('.scroll-mt-20');
            const testePainel = card.querySelector('[data-teste-panel]');
            const snippetPainel = card.querySelector('[data-snippet-panel]');
            card.querySelectorAll('.doc-tab-btn').forEach(b => {
                b.classList.remove('text-[#F2A900]', 'border-[#F2A900]');
                b.classList.add('text-[#8C9B9E]', 'border-transparent');
            });
            btn.classList.add('text-[#F2A900]', 'border-[#F2A900]');
            btn.classList.remove('text-[#8C9B9E]', 'border-transparent');
            if (tab === 'teste') {
                testePainel.classList.remove('hidden');
                snippetPainel.classList.add('hidden');
            } else {
                testePainel.classList.add('hidden');
                snippetPainel.classList.remove('hidden');
                const path = card.dataset.path;
                const method = card.dataset.method;
                const info = (_openapiData && _openapiData.paths && _openapiData.paths[path]) ? _openapiData.paths[path][method.toLowerCase()] : {};
                const snipEl = card.querySelector('[data-snippet-code]');
                const activeLang = card.querySelector('.active-lang');
                const lang = activeLang ? activeLang.textContent.trim().toLowerCase().replace('javascript', 'js') : 'curl';
                if (snipEl) {
                    const code = gerarSnippet(lang, method, path, info, card);
                    snipEl.innerHTML = destacarSintaxe(lang, code);
                }
            }
        }

        function destacarSintaxe(lang, code) {
            if (!code) return '';
            const safe = code.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
            const tokens = [];
            const tokenPlaceholder = function(idx) { return '___TOK_' + idx + '___'; };

            let text = safe.replace(/("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')/g, function(match) {
                const idx = tokens.length;
                tokens.push('<span class="text-[#00A86B]">' + match + '</span>');
                return tokenPlaceholder(idx);
            });

            function empacotar(regex, classe) {
                text = text.replace(regex, function(match) {
                    const idx = tokens.length;
                    tokens.push('<span class="' + classe + '">' + match + '</span>');
                    return tokenPlaceholder(idx);
                });
            }

            if (lang === 'curl') {
                empacotar(/(^|\s)(curl|-X|-H|-d)(?=\s|$)/g, 'text-blue-400 font-bold');
            } else if (lang === 'js') {
                empacotar(/\b(const|let|var|await|async|function|return)\b/g, 'text-purple-400 font-bold');
                empacotar(/\b\d+\b/g, 'text-[#F2A900]');
            } else if (lang === 'python') {
                empacotar(/\b(import|from|def|return|print)\b/g, 'text-purple-400 font-bold');
                empacotar(/\b\d+\b/g, 'text-[#F2A900]');
            } else if (lang === 'harpia') {
                empacotar(/\b(de|importe|var|funcao|retornar|se|enquanto|imprimir|requisitar)\b/g, 'text-[#F2A900] font-bold');
                empacotar(/\b\d+\b/g, 'text-blue-400');
            }

            tokens.forEach(function(tHtml, idx) {
                text = text.replace(tokenPlaceholder(idx), tHtml);
            });

            return text;
        }

        function mudarLang(btn, id, lang) {
            const card = document.getElementById(id);
            card.querySelectorAll('.snippet-lang').forEach(b => {
                b.classList.remove('active-lang', 'bg-[#F2A900]/20', 'text-[#F2A900]', 'font-bold', 'border', 'border-[#F2A900]/20');
                b.classList.add('text-[#8C9B9E]');
            });
            btn.classList.add('active-lang', 'bg-[#F2A900]/20', 'text-[#F2A900]', 'font-bold', 'border', 'border-[#F2A900]/20');
            btn.classList.remove('text-[#8C9B9E]');
            const path = card.dataset.path;
            const method = card.dataset.method;
            const info = (_openapiData && _openapiData.paths && _openapiData.paths[path]) ? _openapiData.paths[path][method.toLowerCase()] : {};
            const snipEl = card.querySelector('[data-snippet-code]');
            if (snipEl) {
                const code = gerarSnippet(lang, method, path, info, card);
                snipEl.innerHTML = destacarSintaxe(lang, code);
            }
        }

        function obterBodyAtual(card, info) {
            const bodyArea = card ? card.querySelector('[data-request-body]') : null;
            if (bodyArea && bodyArea.value.trim()) return bodyArea.value.trim();
            if (info && info.requestBody && info.requestBody.content && info.requestBody.content['application/json']) {
                const cObj = info.requestBody.content['application/json'];
                return JSON.stringify(cObj.example || {}, null, 2);
            }
            return null;
        }

        let _baseUrl = localStorage.getItem('harpia_base_url') || ('http://localhost:' + (location.port || '3000'));
        function alterarBaseUrl(novaUrl) {
            if (!novaUrl) return;
            _baseUrl = novaUrl.trim().replace(/\/+$/, '');
            localStorage.setItem('harpia_base_url', _baseUrl);
            document.querySelectorAll('#endpoints-container > div').forEach(card => {
                const snipEl = card.querySelector('[data-snippet-code]');
                const activeLang = card.querySelector('.active-lang');
                const lang = activeLang ? activeLang.textContent.trim().toLowerCase().replace('javascript', 'js') : 'curl';
                const path = card.dataset.path;
                const method = card.dataset.method;
                const info = (_openapiData && _openapiData.paths && _openapiData.paths[path]) ? _openapiData.paths[path][method.toLowerCase()] : {};
                if (snipEl) snipEl.innerHTML = destacarSintaxe(lang, gerarSnippet(lang, method, path, info, card));
            });
        }

        function gerarSnippet(lang, method, path, info, card) {
            const url = _baseUrl + path.replace(/:([a-zA-Z0-9_]+)/g, '{$1}');
            const body = obterBodyAtual(card, info);
            const authHeader = token ? '  -H "Authorization: Bearer ' + token.slice(0,20) + '..."' : '';
            if (lang === 'curl') {
                let s = 'curl -X ' + (method || 'GET').toUpperCase() + ' "' + url + '" \\';
                s += '\n  -H "Content-Type: application/json"';
                if (authHeader) s += ' \\'  + '\n' + authHeader;
                if (body) s += ' \\'  + '\n  -d \'' + body.replace(/\n/g, '\n       ') + '\'';
                return s;
            }
            if (lang === 'js') {
                let s = 'const resp = await fetch("' + url + '", {';
                s += '\n  method: "' + (method || 'GET').toUpperCase() + '",';
                s += '\n  headers: { "Content-Type": "application/json"' + (token ? ', "Authorization": "Bearer TOKEN"' : '') + ' },';
                if (body) s += '\n  body: JSON.stringify(' + body + ')';
                s += '\n});\nconst data = await resp.json();';
                return s;
            }
            if (lang === 'python') {
                let s = 'import requests\n\n';
                s += 'resp = requests.' + (method || 'get').toLowerCase() + '(\n';
                s += '    "' + url + '",\n';
                s += '    headers={"Content-Type": "application/json"' + (token ? ', "Authorization": "Bearer TOKEN"' : '') + '}';
                if (body) s += ',\n    json=' + body.replace(/"/g, "'");
                s += '\n)\nprint(resp.json())';
                return s;
            }
            if (lang === 'harpia') {
                let s = 'de "http" importe requisitar\n\n';
                if (body) s += 'var corpo = ' + body + '\n\n';
                s += 'var resposta = requisitar("' + (method || 'GET').toUpperCase() + '", "' + url + '"';
                if (body) s += ', corpo';
                s += ')\nimprimir(resposta)';
                return s;
            }
            return '';
        }

        function copiarSnippet(btn) {
            const snipEl = btn.closest('[data-snippet-panel]').querySelector('[data-snippet-code]');
            if (!snipEl) return;
            navigator.clipboard.writeText(snipEl.textContent).then(() => {
                const orig = btn.textContent;
                btn.textContent = '✓ Copiado!';
                btn.style.color = '#00A86B';
                setTimeout(() => { btn.textContent = orig; btn.style.color = ''; }, 1500);
            });
        }

        // ===== HISTÓRICO =====
        function carregarHistorico() {
            try { return JSON.parse(localStorage.getItem('harpia_historico') || '[]'); } catch { return []; }
        }
        function salvarHistoricoLS(hist) {
            localStorage.setItem('harpia_historico', JSON.stringify(hist.slice(-20)));
        }
        function renderizarHistorico() {
            const hist = carregarHistorico();
            const c = document.getElementById('historico-container');
            if (!c) return;
            if (hist.length === 0) { c.innerHTML = '<p class="text-[10px] text-slate-700 font-mono-harpia italic">Sem chamadas salvas.</p>'; return; }
            c.innerHTML = hist.slice().reverse().map((h, i) => {
                const idx = hist.length - 1 - i;
                return '<button onclick="restaurarHistorico(' + idx + ')" class="w-full text-left group flex flex-col gap-0.5 py-1.5 px-2 rounded-lg hover:bg-[#171E26]/40 border border-transparent hover:border-slate-800/50 transition">' +
                    '<span class="text-[10px] font-bold font-mono-harpia text-slate-400 group-hover:text-[#F2A900] truncate">' + (h.method || 'POST') + ' ' + (h.path || '') + '</span>' +
                    '<span class="text-[9px] text-slate-700 font-mono-harpia truncate">' + new Date(h.ts).toLocaleTimeString('pt-BR') + '</span>' +
                    '</button>';
            }).join('');
        }
        function salvarNoHistorico(btn) {
            const card = btn.closest('.scroll-mt-20');
            const path = card.dataset.path;
            const method = card.dataset.method;
            const bodyArea = card.querySelector('[data-request-body]');
            const params = {};
            card.querySelectorAll('[data-route-param]').forEach(inp => { params[inp.dataset.routeParam] = inp.value; });
            const hist = carregarHistorico();
            hist.push({ path, method, body: bodyArea ? bodyArea.value : '', params, ts: Date.now() });
            salvarHistoricoLS(hist);
            renderizarHistorico();
            const orig = btn.textContent;
            btn.textContent = '✓';
            btn.style.color = '#00A86B';
            setTimeout(() => { btn.textContent = orig; btn.style.color = ''; }, 1200);
        }
        function restaurarHistorico(idx) {
            const hist = carregarHistorico();
            const h = hist[idx];
            if (!h) return;
            const id = (h.method || 'post').toLowerCase() + '-' + (h.path || '').replace(/[^a-zA-Z0-9]/g, '-');
            const card = document.getElementById(id);
            if (!card) return;
            card.scrollIntoView({ behavior: 'smooth', block: 'start' });
            const bodyArea = card.querySelector('[data-request-body]');
            if (bodyArea && h.body) bodyArea.value = h.body;
            Object.entries(h.params || {}).forEach(([k, v]) => {
                const inp = card.querySelector('[data-route-param="' + k + '"]');
                if (inp) inp.value = v;
            });
        }
        function limparHistorico() {
            localStorage.removeItem('harpia_historico');
            renderizarHistorico();
        }
        renderizarHistorico();

        // ===== CORRENTEZA =====
        let correntezaModo = 'bossa';
        function mudarCorrenteza(modo) {
            correntezaModo = modo;
        }
        function aplicarCorrenteza() {
            if (correntezaModo === '3g') {
                const delay = Math.random() * 1500 + 1500;
                return new Promise(r => setTimeout(r, delay));
            }
            if (correntezaModo === 'instavel') {
                if (Math.random() < 0.2) {
                    return Promise.reject(new Error('HRP-NET-5005: Instabilidade Tropical — conexão perdida aleatoriamente.'));
                }
            }
            return Promise.resolve();
        }

        // ===== VARIÁVEIS DE AMBIENTE =====
        function carregarVars() {
            try { return JSON.parse(localStorage.getItem('harpia_vars') || '{}'); } catch { return {}; }
        }
        function renderizarVarsModal() {
            const vars = carregarVars();
            const container = document.getElementById('vars-list');
            const keys = Object.keys(vars);
            if (keys.length === 0) {
                container.innerHTML = '<p class="text-xs text-slate-500 italic">Nenhuma variável configurada. Clique em + Adicionar.</p>';
                return;
            }
            container.innerHTML = keys.map(k =>
                '<div class="flex gap-2 items-center var-row">' +
                    '<input type="text" placeholder="NOME" value="' + k + '" class="var-key w-1/3 bg-[#0A0F0D] border border-slate-800 rounded px-2 py-1 text-xs text-[#F2A900] uppercase font-bold" />' +
                    '<input type="text" placeholder="VALOR" value="' + (vars[k] || '') + '" class="var-val w-2/3 bg-[#0A0F0D] border border-slate-800 rounded px-2 py-1 text-xs text-slate-200" />' +
                    '<button onclick="this.parentElement.remove()" class="text-rose-400 text-xs px-1">✕</button>' +
                '</div>'
            ).join('');
        }
        function abrirModalVars() {
            renderizarVarsModal();
            const modal = document.getElementById('modal-vars');
            modal.classList.remove('hidden');
            setTimeout(() => modal.classList.remove('opacity-0'), 10);
        }
        function fecharModalVars() {
            const modal = document.getElementById('modal-vars');
            modal.classList.add('opacity-0');
            setTimeout(() => modal.classList.add('hidden'), 200);
        }
        function adicionarLinhaVar() {
            const container = document.getElementById('vars-list');
            if (container.querySelector('p')) container.innerHTML = '';
            const div = document.createElement('div');
            div.className = 'flex gap-2 items-center var-row';
            div.innerHTML =
                '<input type="text" placeholder="NOME" class="var-key w-1/3 bg-[#0A0F0D] border border-slate-800 rounded px-2 py-1 text-xs text-[#F2A900] uppercase font-bold" />' +
                '<input type="text" placeholder="VALOR" class="var-val w-2/3 bg-[#0A0F0D] border border-slate-800 rounded px-2 py-1 text-xs text-slate-200" />' +
                '<button onclick="this.parentElement.remove()" class="text-rose-400 text-xs px-1">✕</button>';
            container.appendChild(div);
        }
        function salvarVars() {
            const vars = {};
            document.querySelectorAll('.var-row').forEach(row => {
                const k = row.querySelector('.var-key').value.trim().toUpperCase();
                const v = row.querySelector('.var-val').value.trim();
                if (k) vars[k] = v;
            });
            localStorage.setItem('harpia_vars', JSON.stringify(vars));
            fecharModalVars();
        }
        function interpolarVars(text) {
            if (!text) return text;
            const vars = carregarVars();
            let result = text;
            Object.entries(vars).forEach(([k, v]) => {
                const regex = new RegExp('\\{\\{\\s*' + k + '\\s*\\}\\}', 'g');
                result = result.replace(regex, v);
            });
            return result;
        }

        // ===== SPOTLIGHT BUSCA =====
        function abrirSpotlight() {
            const modal = document.getElementById('modal-spotlight');
            modal.classList.remove('hidden');
            setTimeout(() => {
                modal.classList.remove('opacity-0');
                const inp = document.getElementById('spotlight-input');
                inp.value = '';
                inp.focus();
                filtrarSpotlight('');
            }, 10);
        }
        function fecharSpotlight() {
            const modal = document.getElementById('modal-spotlight');
            modal.classList.add('opacity-0');
            setTimeout(() => modal.classList.add('hidden'), 200);
        }
        function filtrarSpotlight(q) {
            const resContainer = document.getElementById('spotlight-results');
            if (!_openapiData || !_openapiData.paths) {
                resContainer.innerHTML = '<p class="text-xs text-slate-500 p-2">Especificação não carregada.</p>';
                return;
            }
            const query = q.toLowerCase().trim();
            const matches = [];
            for (const [path, methods] of Object.entries(_openapiData.paths)) {
                for (const [method, info] of Object.entries(methods)) {
                    const fullText = (method + ' ' + path + ' ' + (info.summary || '') + ' ' + (info.description || '')).toLowerCase();
                    if (!query || fullText.includes(query)) {
                        matches.push({ path, method: method.toUpperCase(), summary: info.summary || path });
                    }
                }
            }
            if (matches.length === 0) {
                resContainer.innerHTML = '<p class="text-xs text-slate-500 p-2 italic">Nenhuma rota encontrada.</p>';
                return;
            }
            resContainer.innerHTML = matches.slice(0, 15).map(m => {
                const id = m.method.toLowerCase() + '-' + m.path.replace(/[^a-zA-Z0-9]/g, '-');
                return '<button onclick="fecharSpotlight(); document.getElementById(\'' + id + '\')?.scrollIntoView({behavior:\'smooth\', block:\'start\'})" class="w-full text-left flex items-center justify-between p-2 rounded hover:bg-[#0A0F0D] border border-transparent hover:border-slate-800 transition">' +
                    '<div class="flex items-center gap-2 truncate">' +
                        '<span class="text-[10px] font-bold px-1.5 py-0.5 rounded bg-slate-800 text-[#F2A900]">' + m.method + '</span>' +
                        '<span class="text-xs text-slate-200 font-bold truncate">' + m.path + '</span>' +
                    '</div>' +
                    '<span class="text-[10px] text-slate-500 truncate max-w-[200px]">' + m.summary + '</span>' +
                '</button>';
            }).join('');
        }

        document.addEventListener('keydown', e => {
            if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
                e.preventDefault();
                abrirSpotlight();
            } else if (e.key === 'Escape') {
                fecharSpotlight();
                fecharModalVars();
            }
        });

        // ===== EXPORTADORES =====
        function exportarPostman() {
            if (!_openapiData) return;
            const items = [];
            for (const [path, methods] of Object.entries(_openapiData.paths || {})) {
                for (const [method, info] of Object.entries(methods)) {
                    items.push({
                        name: info.summary || (method.toUpperCase() + ' ' + path),
                        request: {
                            method: method.toUpperCase(),
                            header: [{ key: "Content-Type", value: "application/json" }],
                            url: { raw: _baseUrl + path, host: [_baseUrl], path: path.split('/').filter(Boolean) }
                        }
                    });
                }
            }
            const postmanColl = {
                info: { name: _openapiData.info.title || "Harpia API", schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" },
                item: items
            };
            downloadBlob(JSON.stringify(postmanColl, null, 2), "harpia_postman_collection.json", "application/json");
        }

        function exportarInsomnia() {
            if (!_openapiData) return;
            const resources = [
                { _id: "wrk_1", _type: "workspace", name: _openapiData.info.title || "Harpia API" }
            ];
            let idx = 1;
            for (const [path, methods] of Object.entries(_openapiData.paths || {})) {
                for (const [method, info] of Object.entries(methods)) {
                    resources.push({
                        _id: "req_" + (idx++),
                        _type: "request",
                        parentId: "wrk_1",
                        name: info.summary || (method.toUpperCase() + ' ' + path),
                        method: method.toUpperCase(),
                        url: _baseUrl + path
                    });
                }
            }
            const insomniaObj = { _type: "export", __export_format: 4, resources };
            downloadBlob(JSON.stringify(insomniaObj, null, 2), "harpia_insomnia.json", "application/json");
        }

        function exportarBruno() {
            if (!_openapiData) return;
            let bruText = "# Bruno Collection Export\n\n";
            for (const [path, methods] of Object.entries(_openapiData.paths || {})) {
                for (const [method, info] of Object.entries(methods)) {
                    bruText += "meta {\n  name: " + (info.summary || path) + "\n  type: http\n}\n\n";
                    bruText += method.toLowerCase() + " {\n  url: " + _baseUrl + path + "\n  body: json\n}\n\n---\n\n";
                }
            }
            downloadBlob(bruText, "harpia_collection.bru", "text/plain");
        }

        function downloadBlob(content, filename, contentType) {
            const blob = new Blob([content], { type: contentType });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            a.click();
            URL.revokeObjectURL(url);
        }

        // ===== SUPER PAINÉIS (Métricas / OWASP / Mentor) =====
        fetch('/docs.json')
            .then(res => res.json())
            .then(data => {
                const mc = document.getElementById('metricas-container');
                let total = 0;
                let metricsHTML = '';
                for (const [path, methods] of Object.entries(data.paths || {})) {
                    for (const [method, info] of Object.entries(methods)) {
                        const count = info['x-harpia-metricas'] || 0;
                        total += count;
                        metricsHTML += '<div class="flex justify-between items-center"><span class="text-[#8C9B9E]">' + method.toUpperCase() + ' ' + path + '</span><span class="text-[#F2A900] font-bold">' + count + '</span></div>';
                    }
                }
                if (total > 0) metricsHTML = '<div class="flex justify-between items-center border-t border-slate-900 pt-1.5 mb-1.5"><span class="text-[#F3F6F4] font-bold">Total</span><span class="text-[#F2A900] font-bold">' + total + '</span></div>' + metricsHTML;
                mc.innerHTML = metricsHTML || '<p class="text-[#8C9B9E] italic">Nenhuma requisição registrada ainda.</p>';

                const ow = document.getElementById('owasp-container');
                const headers = [
                    ['X-Content-Type-Options: nosniff', 'Previne MIME-sniffing'],
                    ['X-Frame-Options: DENY', 'Bloqueia iframe externo (Clickjacking)'],
                    ['X-XSS-Protection: 1; mode=block', 'Filtro XSS do navegador'],
                    ['Referrer-Policy: strict-origin-when-cross-origin', 'Protege dados no referrer'],
                    ['CORS: Access-Control-Allow-Origin: *', 'Suporte nativo a APIs cross-origin']
                ];
                ow.innerHTML = headers.map(h => '<li class="flex items-start gap-2"><span class="text-[#00A86B]">&#10003;</span><div><span class="font-mono-harpia text-[#F3F6F4] text-[10px]">' + h[0] + '</span><br><span class="text-[#8C9B9E] text-[9px]">' + h[1] + '</span></div></li>').join('');

                const ROTAS_MENTOR_SISTEMA = new Set(['/*', '/docs', '/docs.json', '/favicon.ico', '/proxy-poster']);
                const me = document.getElementById('mentor-container');
                const alerts = [];
                let allPaths = [];
                let versionAlertado = false;
                for (const [path, methods] of Object.entries(data.paths || {})) {
                    if (ROTAS_MENTOR_SISTEMA.has(path)) continue;
                    const parts = path.split('/').filter(Boolean);
                    const last = parts[parts.length - 1] || '';
                    allPaths.push({ path: path, methods: methods, parts: parts, last: last });

                    if (/^(criar|salvar|deletar|remover|editar|atualizar|buscar|listar|obter|create|delete|update|get|list|fetch|save|remove|edit|find)$/i.test(last)) {
                        alerts.push('<span class="text-[#D45D34]">&#9888;</span> <span class="text-[#F3F6F4]">' + path + '</span> — verbo de ação no path. Use substantivo.');
                    }

                    if (/[a-z][A-Z]/.test(path)) {
                        alerts.push('<span class="text-[#F2A900]">&#9888;</span> <span class="text-[#F3F6F4]">' + path + '</span> — use kebab-case consistente.');
                    }

                    if (!ROTAS_MENTOR_SISTEMA.has(path)) {
                        const hasVersion = /^\/v\d+/.test(path);
                        if (!hasVersion && versionAlertado === false) {
                            alerts.push('<span class="text-[#8C9B9E]">&#8505;</span> Considere versionar a API com prefixo <code class="text-[#F2A900]">/v1/</code>.');
                            versionAlertado = true;
                        }

                        for (const [m, _info] of Object.entries(methods)) {
                            const ml = m.toLowerCase();
                            if (['post','put','patch','delete'].includes(ml)) {
                                alerts.push('<span class="text-[#D45D34]">&#9888;</span> <span class="text-[#F3F6F4]">' + ml.toUpperCase() + ' ' + path + '</span> — método mutável. Documente auth.');
                            }
                        }

                        if (/:[a-zA-Z_]+/.test(path) && !path.match(/:\w+\([^)]+\)/)) {
                            alerts.push('<span class="text-[#8C9B9E]">&#8505;</span> <span class="text-[#F3F6F4]">' + path + '</span> — adicione constraint ao param (ex: <code class="text-[#F2A900]">:id([0-9]+)</code>).');
                        }
                    }
                }

                if (allPaths.length > 0) {
                    const getCount = allPaths.filter(p => p.methods.get).length;
                    const postCount = allPaths.filter(p => p.methods.post).length;
                    const summary = '<li class="text-[#00A86B] italic border-t border-slate-900 pt-1.5 mt-1.5">&#10003; API com ' + allPaths.length + ' rotas (' + getCount + ' GET, ' + postCount + ' POST).</li>';
                    alerts.push(summary);
                }

                me.innerHTML = alerts.length > 0 ? alerts.map(a => '<li class="flex items-start gap-2">' + a + '</li>').join('') : '<li class="text-[#00A86B] italic">&#10003; Nenhuma heurística violada.</li>';
            })
            .catch(() => {});

        function abrirStandalone() {
            window.open(location.href, '_blank', 'width=1280,height=860,menubar=no,toolbar=no,location=no');
        }

        let falhasConexaoConsecutivas = 0;
        function verificarSaudeServidor() {
            const start = performance.now();
            fetch('/docs.json', { cache: 'no-store' })
                .then(res => {
                    if (!res.ok) throw new Error();
                    const duration = performance.now() - start;
                    falhasConexaoConsecutivas = 0;
                    document.getElementById('banner-servidor-offline').classList.add('hidden');
                    const dot = document.getElementById('status-servidor-dot');
                    const text = document.getElementById('status-servidor-text');
                    if (dot && text) {
                        if (duration < 200) {
                            dot.className = 'w-2 h-2 rounded-full bg-[#00A86B] pulse-online';
                            text.className = 'text-[10px] font-bold text-[#00A86B] font-mono-harpia';
                            text.textContent = 'ONLINE (' + Math.round(duration) + 'ms)';
                        } else {
                            dot.className = 'w-2 h-2 rounded-full bg-[#F2A900]';
                            text.className = 'text-[10px] font-bold text-[#F2A900] font-mono-harpia';
                            text.textContent = 'LENTO (' + Math.round(duration) + 'ms)';
                        }
                    }
                })
                .catch(() => {
                    falhasConexaoConsecutivas++;
                    const dot = document.getElementById('status-servidor-dot');
                    const text = document.getElementById('status-servidor-text');
                    if (dot && text) {
                        dot.className = 'w-2 h-2 rounded-full bg-[#D45D34] pulse-offline';
                        text.className = 'text-[10px] font-bold text-[#D45D34] font-mono-harpia';
                        text.textContent = 'OFFLINE';
                    }
                    if (falhasConexaoConsecutivas >= 3) {
                        const banner = document.getElementById('banner-servidor-offline');
                        banner.classList.remove('hidden');
                        document.getElementById('segundos-offline').textContent = (falhasConexaoConsecutivas * 3);
                    }
                });
        }
        setInterval(verificarSaudeServidor, 3000);

        let filtroMetodoAtivo = 'TODOS';
        function filtrarPorMetodo(metodo) {
            filtroMetodoAtivo = metodo;
            document.querySelectorAll('.filtro-btn').forEach(btn => {
                if (btn.textContent.trim() === metodo) {
                    btn.className = 'filtro-btn active-filtro px-3 py-1 rounded-lg border border-[#F2A900]/30 bg-[#F2A900]/10 text-[#F2A900] font-bold transition';
                } else {
                    btn.className = 'filtro-btn px-3 py-1 rounded-lg border border-slate-800 text-[#8C9B9E] hover:bg-[#171E26]/40 transition';
                }
            });

            document.querySelectorAll('[data-method]').forEach(card => {
                const cardMethod = (card.dataset.method || '').toUpperCase();
                if (metodo === 'TODOS' || cardMethod === metodo) {
                    card.style.display = '';
                } else {
                    card.style.display = 'none';
                }
            });
        }

        let sidebarObserver = null;
        function iniciarObserverSidebar() {
            if (sidebarObserver) sidebarObserver.disconnect();
            sidebarObserver = new IntersectionObserver(entries => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const id = entry.target.id;
                        document.querySelectorAll('.sidebar-rota').forEach(link => {
                            if (link.getAttribute('href') === '#' + id) {
                                link.classList.add('ativa-rota');
                            } else {
                                link.classList.remove('ativa-rota');
                            }
                        });
                    }
                });
            }, { rootMargin: '-15% 0px -60% 0px' });

            document.querySelectorAll('#endpoints-container > div').forEach(el => {
                sidebarObserver.observe(el);
            });
        }

        setTimeout(iniciarObserverSidebar, 300);

        let metricasAnteriores = {};
        function atualizarMetricasComDiff(data) {
            const mc = document.getElementById('metricas-container');
            if (!mc) return;
            let total = 0;
            let html = '';
            const novasMetricas = {};

            for (const [path, methods] of Object.entries(data.paths || {})) {
                for (const [method, info] of Object.entries(methods)) {
                    const count = info['x-harpia-metricas'] || 0;
                    total += count;
                    const key = method.toUpperCase() + ' ' + path;
                    novasMetricas[key] = count;

                    const countAntigo = metricasAnteriores[key] || 0;
                    const mudou = (count > countAntigo && metricasAnteriores[key] !== undefined);
                    const flashClass = mudou ? 'flash-mudanca text-[#00A86B]' : 'text-[#F2A900]';

                    html += '<div class="flex justify-between items-center py-0.5 px-1 rounded transition-colors"><span class="text-[#8C9B9E] truncate">' + key + '</span><span class="font-bold ' + flashClass + '">' + count + (mudou ? ' ▲' : '') + '</span></div>';
                }
            }
            if (total > 0) {
                html = '<div class="flex justify-between items-center border-t border-slate-900 pt-1.5 mb-1.5 font-bold"><span class="text-[#F3F6F4]">Total</span><span class="text-[#F2A900]">' + total + '</span></div>' + html;
            }
            mc.innerHTML = html || '<p class="text-[#8C9B9E] italic">Nenhuma requisição registrada ainda.</p>';
            metricasAnteriores = novasMetricas;
        }

        setInterval(() => {
            fetch('/docs.json').then(r => r.json()).then(data => {
                atualizarMetricasComDiff(data);
            }).catch(() => {});
        }, 5000);

        // ===== ANIMAÇÃO DE SLIDE UP/DOWN E SHRINK DO CABEÇALHO PÍLULA =====
        function atualizarScrollHeader() {
            const header = document.querySelector('header');
            if (!header) return;
            const currentScrollY = window.scrollY;

            // 1. Diminuir o tamanho (shrink) ao rolar > 30px
            if (currentScrollY > 30) {
                header.classList.add('header-shrunk');
            } else {
                header.classList.remove('header-shrunk');
            }

            // 2. Se estiver na parte superior da página (scrollY <= 150), o cabeçalho SEMPRE deve estar visível
            if (currentScrollY <= 150) {
                header.classList.remove('header-hidden');
                return;
            }

            // 3. Ocultar com slide-up ao chegar bem próximo dos filtros de endpoints (#barra-filtros-wrapper)
            const filtrosWrapper = document.getElementById('barra-filtros-wrapper');
            if (filtrosWrapper && filtrosWrapper.offsetParent !== null) {
                const rect = filtrosWrapper.getBoundingClientRect();
                if (rect.top < 25 || currentScrollY > 450) {
                    header.classList.add('header-hidden');
                } else {
                    header.classList.remove('header-hidden');
                }
            } else if (currentScrollY > 450) {
                header.classList.add('header-hidden');
            } else {
                header.classList.remove('header-hidden');
            }
        }

        window.addEventListener('scroll', atualizarScrollHeader, { passive: true });
        window.addEventListener('DOMContentLoaded', () => {
            requestAnimationFrame(() => atualizarScrollHeader());
        });
        setTimeout(atualizarScrollHeader, 100);
    </script>
</body>
</html>`
