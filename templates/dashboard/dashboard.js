let groups = [];
let urls = [];
let selectedGroupId = null;
const token = localStorage.getItem('token');

if (!token) window.location.href = '/login';

const headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
};

async function apiFetch(url, options = {}) {
    const res = await fetch(url, options);
    if (res.status === 401) {
        localStorage.clear();
        window.location.href = '/login';
        throw new Error('Unauthorized');
    }
    return res;
}

async function init() {
    await fetchGroups();
}

async function fetchGroups() {
    try {
        const res = await apiFetch('/api/groups', { headers });
        groups = await res.json() || [];
        renderDropdown();
        
        if (groups.length > 0 && !selectedGroupId) {
            selectGroup(groups[0].id, groups[0].name);
        } else if (groups.length === 0) {
            document.getElementById('selectedGroupText').innerText = 'Сначала создайте группу...';
            document.getElementById('urlsTableBody').innerHTML = '';
        }
    } catch (e) { console.error(e); }
}

async function createGroup() {
    const name = document.getElementById('groupNameInput').value.trim();
    if (!name) return alert('Введите название группы');

    try {
        const res = await apiFetch('/api/groups', {
            method: 'POST',
            headers,
            body: JSON.stringify({ name })
        });
        if (res.ok) {
            document.getElementById('groupNameInput').value = '';
            await fetchGroups();
        } else { alert('Ошибка создания'); }
    } catch (e) { console.error(e); }
}

async function deleteGroup(id) {
    if (!confirm('Удалить группу и все её товары?')) return;
    try {
        const res = await apiFetch(`/api/groups/${id}`, { method: 'DELETE', headers });
        if (res.ok) {
            if (selectedGroupId === id) selectedGroupId = null;
            await fetchGroups();
        }
    } catch (e) { console.error(e); }
}

function renderDropdown() {
    const dropdown = document.getElementById('dropdownContent');
    dropdown.innerHTML = '';
    if (groups.length === 0) {
        dropdown.innerHTML = '<div style="padding: 12px; color: #64748b; text-align: center;">Нет групп</div>';
        return;
    }
    groups.forEach(g => {
        const item = document.createElement('div');
        item.className = 'dropdown-item';
        item.innerHTML = `
            <span class="group-name" onclick="selectGroup(${g.id}, '${g.name}')">${g.name}</span>
            <button class="btn-icon danger" onclick="event.stopPropagation(); deleteGroup(${g.id})">✕</button>
        `;
        dropdown.appendChild(item);
    });
}

function toggleDropdown(event) {
    if (event.target.tagName === 'BUTTON' || event.target.closest('button')) return;
    document.getElementById('customSelect').classList.toggle('active');
}

window.onclick = (e) => {
    if (!e.target.closest('.custom-select')) document.getElementById('customSelect').classList.remove('active');
};

function selectGroup(id, name) {
    selectedGroupId = id;
    document.getElementById('selectedGroupText').innerText = name;
    document.getElementById('customSelect').classList.remove('active');
    fetchUrls();
}

async function fetchUrls() {
    if (!selectedGroupId) return;
    try {
        const res = await apiFetch(`/api/urls?group_id=${selectedGroupId}`, { headers });
        urls = await res.json() || [];
        renderTable();
    } catch (e) { console.error(e); }
}

async function addUrl() {
    const url = document.getElementById('urlInput').value.trim();
    if (!selectedGroupId) return alert('Выберите группу');
    if (!url) return alert('Введите URL');

    try {
        const res = await apiFetch('/api/urls', {
            method: 'POST',
            headers,
            body: JSON.stringify({ url: url, group_id: selectedGroupId })
        });
        if (res.ok) {
            document.getElementById('urlInput').value = '';
            await fetchUrls();
        } else { alert('Ошибка добавления URL'); }
    } catch (e) { console.error(e); }
}

async function deleteUrl(id) {
    if (!confirm('Удалить этот товар?')) return;
    try {
        const res = await apiFetch(`/api/urls/${id}?group_id=${selectedGroupId}`, { method: 'DELETE', headers });
        if (res.ok) await fetchUrls();
    } catch (e) { console.error(e); }
}

async function toggleActive(id) {
    try {
        const res = await apiFetch(`/api/urls/${id}/active`, { method: 'PATCH', headers });
        if (res.ok) await fetchUrls();
    } catch (e) { console.error(e); }
}

function renderTable() {
    const tbody = document.getElementById('urlsTableBody');
    tbody.innerHTML = '';
    
    const group = groups.find(g => g.id === selectedGroupId);
    const groupName = group ? group.name : 'Неизвестно';

    urls.forEach(u => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td><span class="badge">${groupName}</span></td>
            <td style="word-break: break-all; color: ${u.active ? 'var(--text-main)' : 'var(--text-muted)'};">
                ${u.url}
            </td>
            <td style="text-align: right;">
                <button class="btn-icon ${u.active ? 'success' : ''}" onclick="toggleActive(${u.id})" title="${u.active ? 'Выключить' : 'Включить'}">
                    ${u.active ? 'Вкл' : 'Выкл'}
                </button>
                <button class="btn-icon danger" onclick="deleteUrl(${u.id})" title="Удалить">✕</button>
            </td>
        `;
        tbody.appendChild(tr);
    });

    if (urls.length === 0) {
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center; color:#94a3b8;">Нет добавленных товаров</td></tr>';
    }
}

function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login';
}

init();