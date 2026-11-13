Chart.defaults.color = '#64748b';
Chart.defaults.borderColor = 'rgba(0,0,0,0.05)';
Chart.defaults.font.family = "'Plus Jakarta Sans', sans-serif";

let historyChartInstance = null;
let sellersChartInstance = null;
let currentAnalyticsData = null;

const MAX_CHART_POINTS = 20;

const priceFormatter = new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
});

const token = localStorage.getItem('token');
if (!token) window.location.href = '/login';

const headers = { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token };

async function apiFetch(url, options = {}) {
    const res = await fetch(url, options);
    if (res.status === 401) {
        localStorage.clear();
        window.location.href = '/login';
        throw new Error('Unauthorized');
    }
    return res;
}

function formatPrice(kopecks) {
    const num = Number(kopecks);
    if (isNaN(num)) return '—';
    return priceFormatter.format(num / 100);
}

function formatDate(dateString) {
    if (!dateString) return '—';
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit'
    });
}

function formatTime(dateString) {
    if (!dateString) return '—';
    return new Date(dateString).toLocaleTimeString('ru-RU', {
        hour: '2-digit', minute: '2-digit'
    });
}

window.onload = async function() {
    await fetchGroups();
};

async function fetchGroups() {
    try {
        const res = await apiFetch('/api/groups', { headers });
        const groups = await res.json() || [];
        
        const select = document.getElementById('groupSelect');
        select.innerHTML = '';
        if (groups.length === 0) {
            select.innerHTML = '<option value="" disabled selected>Нет групп</option>';
            return;
        }
        
        groups.forEach(g => {
            const opt = document.createElement('option');
            opt.value = g.id;
            opt.text = g.name;
            select.appendChild(opt);
        });
        
        select.value = groups[0].id;
        await onGroupChange();
    } catch (e) { console.error(e); }
}

async function onGroupChange() {
    const groupId = document.getElementById('groupSelect').value;
    const urlSelect = document.getElementById('urlSelect');
    urlSelect.innerHTML = '<option value="" disabled selected>Загрузка...</option>';
    
    try {
        const res = await apiFetch(`/api/urls?group_id=${groupId}`, { headers });
        const urls = await res.json() || [];
        
        urlSelect.innerHTML = '';
        if (urls.length === 0) {
            urlSelect.innerHTML = '<option value="" disabled selected>Нет товаров в группе</option>';
        } else {
            urls.forEach(u => {
                const opt = document.createElement('option');
                opt.value = u.id;
                opt.text = u.name;
                urlSelect.appendChild(opt);
            });
            
            urlSelect.value = urls[0].id;
            await loadAnalytics(urls[0].id);
        }
        await loadGroupComparison(groupId);
    } catch (e) { console.error(e); }
}

async function loadAnalytics(idUrl) {
    try {
        const res = await apiFetch(`/api/analytics/${idUrl}`, { headers });
        if (!res.ok) return;
        const data = await res.json();
        currentAnalyticsData = data; 

        document.getElementById('currentPrice').innerText = data.current_price ? formatPrice(data.current_price) : '—';
        
        let min = Infinity, max = -Infinity;
        if (data.history && data.history.length > 0) {
            data.history.forEach(h => {
                const p = parseFloat(h.price);
                if (!isNaN(p)) {
                    if (p < min) min = p;
                    if (p > max) max = p;
                }
            });
        }
        
        document.getElementById('minPrice').innerText = min !== Infinity ? formatPrice(min) : '—';
        document.getElementById('maxPrice').innerText = max !== -Infinity ? formatPrice(max) : '—';
        document.getElementById('sellersCount').innerText = data.seller_name || 'Скрыт';
        document.getElementById('productDescription').innerText = data.description || 'Описание отсутствует';

        const optionsList = document.getElementById('productOptions');
        optionsList.innerHTML = '';
        if (data.options && data.options.length > 0) {
            data.options.forEach(opt => {
                const li = document.createElement('li');
                li.innerHTML = `<strong>${opt.name || 'Свойство'}:</strong> ${opt.value || '—'}`;
                optionsList.appendChild(li);
            });
        } else {
            optionsList.innerHTML = '<li>Характеристики отсутствуют</li>';
        }
        
        renderHistoryChart(data.history || []);

        const alertsBody = document.getElementById('alertsBody');
        alertsBody.innerHTML = '';

        if (data.history && data.history.length > 1) {
            for (let i = data.history.length - 1; i > 0; i--) {
                const current = data.history[i];
                const prev = data.history[i - 1];
                
                let currentPrice = parseFloat(current.price);
                let prevPrice = parseFloat(prev.price);

                if (currentPrice !== prevPrice && !isNaN(currentPrice) && !isNaN(prevPrice)) {
                    const isDrop = currentPrice < prevPrice;
                    const badgeClass = isDrop ? 'success' : 'danger';
                    const badgeText = isDrop ? 'Цена упала' : 'Цена выросла';

                    const tr = document.createElement('tr');
                    tr.innerHTML = `
                        <td><span class="badge ${badgeClass}">${badgeText}</span></td>
                        <td>${data.seller_name || '—'}</td>
                        <td>${formatPrice(currentPrice)} <span style="color:#8b949e;text-decoration:line-through;font-size:12px;">${formatPrice(prevPrice)}</span></td>
                        <td style="color:#8b949e;">${formatDate(current.date)}</td>
                    `;
                    alertsBody.appendChild(tr);
                    if (alertsBody.children.length >= 5) break; 
                }
            }
        }

        if (alertsBody.children.length === 0) {
            alertsBody.innerHTML = '<tr><td colspan="4" style="text-align:center; color:#94a3b8;">Нет изменений цен</td></tr>';
        }

    } catch (e) { console.error(e); }
}

async function loadGroupComparison(groupId) {
    try {
        const res = await apiFetch(`/api/analytics/group/${groupId}`, { headers });
        if (!res.ok) return;
        const data = await res.json() || [];
        
        if (sellersChartInstance) sellersChartInstance.destroy();
        
        const labels = data.map(item => item.name || 'Без имени');
        const prices = data.map(item => (parseFloat(item.price) || 0) / 100);

        const ctx = document.getElementById('sellersChart').getContext('2d');
        sellersChartInstance = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{ 
                    label: 'Текущая цена (₽)', 
                    data: prices, 
                    backgroundColor: '#6366f1', 
                    borderRadius: 6 
                }]
            },
            options: { 
                responsive: true, maintainAspectRatio: false, 
                plugins: { legend: { display: false } },
                scales: { 
                    x: { grid: { display: false }, ticks: { autoSkip: false, maxRotation: 45, minRotation: 45 } }, 
                    y: { grid: { color: 'rgba(0,0,0,0.05)' } } 
                }
            }
        });
    } catch (e) { console.error(e); }
}

function renderHistoryChart(history) {
    if (historyChartInstance) historyChartInstance.destroy();
    
    const dateEl = document.getElementById('chartDate');
    if (dateEl) {
        dateEl.innerText = new Date().toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
    }

    let chartData = history || [];
    if (chartData.length > MAX_CHART_POINTS) {
        chartData = chartData.slice(-MAX_CHART_POINTS);
    }
    if (chartData.length === 0) {
        chartData = [{ date: new Date().toISOString(), price: "0" }];
    }

    const labels = chartData.map(h => formatTime(h.date));
    const prices = chartData.map(h => (parseFloat(h.price) || 0) / 100);
    
    const ctx = document.getElementById('historyChart').getContext('2d');
    historyChartInstance = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Цена (₽)',
                data: prices,
                borderColor: '#6366f1',
                backgroundColor: 'rgba(99, 102, 241, 0.1)',
                borderWidth: 3,
                pointBackgroundColor: '#ffffff',
                pointBorderColor: '#6366f1',
                pointBorderWidth: 2,
                pointRadius: 4,
                fill: true,
                tension: 0.4
            }]
        },
        options: {
            responsive: true, maintainAspectRatio: false,
            plugins: { legend: { display: false } },
            scales: { x: { grid: { display: false } }, y: { grid: { color: 'rgba(0,0,0,0.05)' } } }
        }
    });
}

function openSellersModal() {
    const tbody = document.getElementById('sellersTableBody');
    tbody.innerHTML = '';

    if (!currentAnalyticsData || !currentAnalyticsData.seller_name || currentAnalyticsData.seller_name === '-') {
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center; color:#94a3b8;">Нет данных о продавце</td></tr>';
    } else {
        const tr = document.createElement('tr');
        tr.classList.add('best-price');
        tr.innerHTML = `
            <td>${currentAnalyticsData.seller_name} <span class="best-badge">Текущий</span></td>
            <td style="color:#10b981;font-weight:bold;">${formatPrice(currentAnalyticsData.current_price)}</td>
            <td class="rating">${currentAnalyticsData.rating || '—'} ⭐</td>
        `;
        tbody.appendChild(tr);
    }
    document.getElementById('sellersModal').style.display = 'flex';
}

function closeSellersModal() {
    document.getElementById('sellersModal').style.display = 'none';
}

window.onclick = function(event) {
    if (event.target === document.getElementById('sellersModal')) closeSellersModal();
}