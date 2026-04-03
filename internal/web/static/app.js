function dashboard() {
    return {
        tab: 'overview',
        wsConnected: false,
        events: [],
        sessions: [],
        tasks: [],
        memories: [],
        skills: [],
        stats: { sessions: 0, tasks: 0, memories: 0, skills: 0 },
        showTaskForm: false,
        newTaskDesc: '',
        newTaskCategory: 'general',
        memorySearch: '',
        settings: {
            model_default: '',
            coding_model: '',
            max_l0_items: 20,
            token_budget: 8000,
            auto_compress: true,
            opencode_binary: '',
            opencode_timeout: 30,
        },

        init() {
            this.loadStats();
            this.loadSessions();
            this.loadTasks();
            this.loadMemories();
            this.loadSkills();
            this.loadSettings();
            this.connectWebSocket();
        },

        connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const ws = new WebSocket(protocol + '//' + window.location.host + '/api/events/ws');
            ws.onopen = () => { this.wsConnected = true; };
            ws.onclose = () => { this.wsConnected = false; setTimeout(() => this.connectWebSocket(), 3000); };
            ws.onerror = () => { this.wsConnected = false; };
            ws.onmessage = (e) => {
                try {
                    const event = JSON.parse(e.data);
                    this.events.push(event);
                    if (this.events.length > 200) this.events = this.events.slice(-200);
                    if (event.type === 'session' || event.type === 'task') {
                        this.loadStats();
                        this.loadSessions();
                        this.loadTasks();
                    }
                } catch (err) {}
            };
        },

        async loadStats() {
            try {
                const [s, t, m, sk] = await Promise.all([
                    fetch('/api/sessions').then(r => r.json()),
                    fetch('/api/tasks').then(r => r.json()),
                    fetch('/api/memory').then(r => r.json()),
                    fetch('/api/skills').then(r => r.json()),
                ]);
                this.stats = {
                    sessions: Array.isArray(s) ? s.length : 0,
                    tasks: Array.isArray(t) ? t.length : 0,
                    memories: Array.isArray(m) ? m.length : 0,
                    skills: Array.isArray(sk) ? sk.length : 0,
                };
            } catch (e) {}
        },

        async loadSessions() {
            try {
                const res = await fetch('/api/sessions');
                this.sessions = await res.json();
            } catch (e) {}
        },

        async loadTasks() {
            try {
                const res = await fetch('/api/tasks');
                this.tasks = await res.json();
            } catch (e) {}
        },

        async loadMemories() {
            try {
                const res = await fetch('/api/memory');
                this.memories = await res.json();
            } catch (e) {}
        },

        async loadSkills() {
            try {
                const res = await fetch('/api/skills');
                this.skills = await res.json();
            } catch (e) {}
        },

        async loadSettings() {
            try {
                const res = await fetch('/api/config');
                const cfg = await res.json();
                if (cfg.model) this.settings.model_default = cfg.model.default || '';
                if (cfg.model) this.settings.coding_model = cfg.model.coding || '';
                if (cfg.memory) {
                    this.settings.max_l0_items = cfg.memory.max_l0_items || 20;
                    this.settings.token_budget = cfg.memory.token_budget || 8000;
                    this.settings.auto_compress = cfg.memory.auto_compress !== false;
                }
                if (cfg.opencode) {
                    this.settings.opencode_binary = cfg.opencode.binary || '';
                    this.settings.opencode_timeout = cfg.opencode.timeout || 30;
                }
            } catch (e) {}
        },

        async saveSettings() {
            try {
                await fetch('/api/config', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        model: { default: this.settings.model_default, coding: this.settings.coding_model },
                        memory: { max_l0_items: this.settings.max_l0_items, token_budget: this.settings.token_budget, auto_compress: this.settings.auto_compress },
                        opencode: { binary: this.settings.opencode_binary, timeout: this.settings.opencode_timeout },
                    }),
                });
            } catch (e) {}
        },

        async createTask() {
            if (!this.newTaskDesc) return;
            try {
                await fetch('/api/tasks', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ description: this.newTaskDesc, category: this.newTaskCategory }),
                });
                this.newTaskDesc = '';
                this.showTaskForm = false;
                this.loadTasks();
                this.loadStats();
            } catch (e) {}
        },

        async newSession() {
            try {
                await fetch('/api/sessions', { method: 'POST' });
                this.loadSessions();
                this.loadStats();
            } catch (e) {}
        },

        async viewSession(id) {
            this.tab = 'sessions';
        },

        async deleteSession(id) {
            if (!confirm('Delete session ' + id + '?')) return;
            try {
                await fetch('/api/sessions/' + id, { method: 'DELETE' });
                this.loadSessions();
                this.loadStats();
            } catch (e) {}
        },

        async searchMemories() {
            if (!this.memorySearch) {
                this.loadMemories();
                return;
            }
            try {
                const res = await fetch('/api/memory/search?q=' + encodeURIComponent(this.memorySearch));
                this.memories = await res.json();
            } catch (e) {}
        },

        formatTime(ts) {
            if (!ts) return '';
            const d = new Date(ts);
            return d.toLocaleTimeString();
        },

        formatDate(ts) {
            if (!ts) return '';
            const d = new Date(ts);
            return d.toLocaleDateString() + ' ' + d.toLocaleTimeString();
        },
    };
}
