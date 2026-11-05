// Адрес API. Если страница отдаётся тем же Go-сервером, оставь пустым.
const API = "";

// Сколько дней показывать в полоске занятости
const CALENDAR_DAYS = 14;

// ================= Прогресс эндпоинтов =================

const endpoints = [
  { key: "GET /rooms",               label: "список комнат" },
  { key: "POST /rooms",              label: "создать комнату" },
  { key: "GET /bookings",            label: "брони комнаты" },
  { key: "POST /bookings",           label: "создать бронь" },
  { key: "DELETE /bookings/{id}",    label: "отменить бронь" },
  { key: "GET /rooms/available",     label: "свободные комнаты" },
];
const status = {};

function renderEndpoints() {
  document.getElementById("endpoints").innerHTML = endpoints.map(e => {
    const s = status[e.key];
    const cls = s === true ? "ok" : s === false ? "fail" : "";
    return `<li><span class="dot ${cls}"></span><code>${e.key}</code><span class="label">${e.label}</span></li>`;
  }).join("");

  const done = endpoints.filter(e => status[e.key] === true).length;
  document.getElementById("progressCount").textContent = `${done} из ${endpoints.length}`;
  document.getElementById("progressBar").style.width = `${(done / endpoints.length) * 100}%`;
}

function mark(key, ok) {
  status[key] = ok;
  renderEndpoints();
}

// ================= Запросы =================

async function call(method, path, body) {
  const res = await fetch(API + path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = null; }
  return { ok: res.ok, status: res.status, data };
}

// Ответы 4xx — это правильная работа API (валидация, конфликт),
// а 404/405 на сам маршрут значат, что эндпоинт ещё не написан.
function routeWorks(r) {
  if (r.ok) return true;
  if (r.status === 405) return false;
  if (r.status === 404 && !(r.data && r.data.error)) return false;
  return r.status < 500;
}

// ================= Даты =================
// Работаем с датами в локальном времени, без toISOString,
// иначе в часовых поясах восточнее UTC дата может «уехать» на день назад.

function toISO(d) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}
function fromISO(s) {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}
function addDays(d, n) {
  const x = new Date(d);
  x.setDate(x.getDate() + n);
  return x;
}
function today() {
  const d = new Date();
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}
function formatDate(s) {
  const d = fromISO(s);
  if (isNaN(d)) return s;
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}
function nightsBetween(a, b) {
  return Math.round((fromISO(b) - fromISO(a)) / 86400000);
}
function plural(n, one, few, many) {
  const m10 = n % 10, m100 = n % 100;
  if (m10 === 1 && m100 !== 11) return one;
  if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return few;
  return many;
}
function nightsText(n) {
  return `${n} ${plural(n, "ночь", "ночи", "ночей")}`;
}

const escapeHtml = s => String(s).replace(/[&<>"']/g, c => (
  { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
));

// ================= Уведомления =================

function toast(text, type = "ok") {
  const box = document.getElementById("toasts");
  const el = document.createElement("div");
  el.className = `toast ${type}`;
  el.textContent = text;
  box.appendChild(el);
  setTimeout(() => {
    el.classList.add("hide");
    setTimeout(() => el.remove(), 300);
  }, 3200);
}

// ================= Состояние =================

const state = {
  rooms: [],
  selectedRoom: null,
  bookings: [],
};

const $ = id => document.getElementById(id);
const startInput = $("startAt");
const endInput = $("endAt");
const roomInput = $("roomId");

// ================= Сервер =================

async function checkServer() {
  const pill = $("serverPill");
  try {
    await fetch(API + "/rooms", { method: "GET" });
    pill.className = "pill ok";
    pill.lastElementChild.textContent = "сервер на связи";
  } catch {
    pill.className = "pill fail";
    pill.lastElementChild.textContent = "сервер не отвечает";
  }
}

// ================= Комнаты =================

async function loadRooms() {
  const box = $("rooms");
  try {
    const r = await call("GET", "/rooms");
    if (!r.ok || !Array.isArray(r.data)) throw new Error();
    mark("GET /rooms", true);
    state.rooms = r.data;
    renderRooms();
  } catch {
    mark("GET /rooms", false);
    box.innerHTML = `<div class="hint" style="grid-column: 1 / -1">
      Эндпоинт <code>GET /rooms</code> пока не отвечает. ID комнаты можно ввести в форме вручную.
    </div>`;
  }
}

function renderRooms() {
  const box = $("rooms");
  if (state.rooms.length === 0) {
    box.innerHTML = `<p class="empty" style="grid-column: 1 / -1">Комнат пока нет. Добавь первую ниже.</p>`;
    return;
  }
  box.innerHTML = state.rooms.map(room => `
    <button type="button" class="room ${room.id === state.selectedRoom ? "active" : ""}" data-id="${room.id}">
      <div class="num">№ ${room.id}</div>
      <div class="name">${escapeHtml(room.name ?? "Комната " + room.id)}</div>
    </button>
  `).join("");
  box.querySelectorAll(".room").forEach(b =>
    b.addEventListener("click", () => selectRoom(Number(b.dataset.id)))
  );
}

function roomName(id) {
  const room = state.rooms.find(r => r.id === id);
  return room && room.name ? room.name : `№ ${id}`;
}

function selectRoom(id) {
  state.selectedRoom = id;
  roomInput.value = id;
  renderRooms();
  loadBookings();
}

$("reloadRooms").addEventListener("click", loadRooms);

$("roomForm").addEventListener("submit", async e => {
  e.preventDefault();
  const name = $("roomName").value.trim();
  if (!name) return;
  try {
    const r = await call("POST", "/rooms", { name });
    mark("POST /rooms", routeWorks(r));
    if (!r.ok) {
      toast(r.data?.error || `Не удалось создать комнату (код ${r.status})`, "err");
      return;
    }
    $("roomName").value = "";
    toast(`Комната «${name}» добавлена`);
    await loadRooms();
    if (r.data?.id) selectRoom(r.data.id);
  } catch {
    mark("POST /rooms", false);
    toast("Сервер не отвечает", "err");
  }
});

// ================= Брони =================

async function loadBookings() {
  const box = $("bookings");
  $("bookingsTitle").textContent = state.selectedRoom
    ? `Занятость: ${roomName(state.selectedRoom)}`
    : "Занятость";

  if (!state.selectedRoom) {
    state.bookings = [];
    renderCalendar();
    return;
  }

  try {
    const r = await call("GET", `/bookings?room_id=${state.selectedRoom}`);
    if (!r.ok || !Array.isArray(r.data)) throw new Error();
    mark("GET /bookings", true);
    state.bookings = [...r.data].sort((a, b) => a.start_at.localeCompare(b.start_at));
    renderBookings();
  } catch {
    mark("GET /bookings", false);
    state.bookings = [];
    box.innerHTML = `<div class="hint" style="margin-top:10px">
      Эндпоинт <code>GET /bookings?room_id=…</code> пока не отвечает, поэтому занятость не видна.
    </div>`;
  }
  renderCalendar();
}

function renderBookings() {
  const box = $("bookings");
  if (state.bookings.length === 0) {
    box.innerHTML = `<p class="empty">Броней нет, комната свободна.</p>`;
    return;
  }
  box.innerHTML = state.bookings.map(b => `
    <div class="booking">
      <div>
        <div class="dates">${formatDate(b.start_at)} → ${formatDate(b.end_at)}</div>
        <div class="meta">${nightsText(nightsBetween(b.start_at, b.end_at))} · бронь № ${b.id}</div>
      </div>
      <button type="button" class="btn-del" data-id="${b.id}">Отменить</button>
    </div>
  `).join("");
  box.querySelectorAll(".btn-del").forEach(btn =>
    btn.addEventListener("click", () => cancelBooking(btn.dataset.id))
  );
}

async function cancelBooking(id) {
  if (!confirm(`Отменить бронь № ${id}?`)) return;
  try {
    const r = await call("DELETE", `/bookings/${id}`);
    mark("DELETE /bookings/{id}", routeWorks(r));
    if (!r.ok) {
      toast(r.data?.error || `Не удалось отменить (код ${r.status})`, "err");
      return;
    }
    toast(`Бронь № ${id} отменена`);
    loadBookings();
  } catch {
    mark("DELETE /bookings/{id}", false);
    toast("Сервер не отвечает", "err");
  }
}

// ================= Календарь =================

// День занят, если какая-то бронь начинается не позже него и заканчивается после него
// (день выезда считается свободным — как в отелях).
function isBusy(dayISO) {
  return state.bookings.some(b => b.start_at <= dayISO && dayISO < b.end_at);
}

function renderCalendar() {
  const cal = $("calendar");
  const start = today();
  const end = addDays(start, CALENDAR_DAYS - 1);
  $("calendarRange").textContent = `${formatDate(toISO(start))} — ${formatDate(toISO(end))}`;

  const pickStart = startInput.value;
  const pickEnd = endInput.value;
  const noRoom = !state.selectedRoom;

  let html = "";
  for (let i = 0; i < CALENDAR_DAYS; i++) {
    const d = addDays(start, i);
    const iso = toISO(d);
    const busy = !noRoom && isBusy(iso);
    const picked = pickStart && pickEnd && pickStart <= iso && iso < pickEnd;
    const cls = [
      "day",
      busy ? "busy" : "",
      picked ? "pick" : "",
      i === 0 ? "today" : "",
      noRoom ? "disabled" : "",
    ].join(" ");
    const wd = d.toLocaleDateString("ru-RU", { weekday: "short" });
    html += `<button type="button" class="${cls}" data-date="${iso}" ${busy || noRoom ? "disabled" : ""}
      title="${busy ? "занято" : "нажми, чтобы выбрать"}">
      <span class="wd">${wd}</span><span class="dn">${d.getDate()}</span>
    </button>`;
  }
  cal.innerHTML = html;
  cal.querySelectorAll(".day:not([disabled])").forEach(el =>
    el.addEventListener("click", () => pickDay(el.dataset.date))
  );
}

// Клик по дням: первый клик — заезд, второй — последняя ночь.
let pickingEnd = false;

function pickDay(iso) {
  if (!pickingEnd || iso <= startInput.value) {
    startInput.value = iso;
    endInput.value = toISO(addDays(fromISO(iso), 1));
    pickingEnd = true;
  } else {
    endInput.value = toISO(addDays(fromISO(iso), 1));
    pickingEnd = false;
  }
  updateSummary();
  renderCalendar();
}

// ================= Форма брони =================

function updateSummary() {
  const s = startInput.value, e = endInput.value;
  const el = $("summary");
  if (!s || !e) { el.textContent = ""; return; }
  const n = nightsBetween(s, e);
  if (n <= 0) {
    el.textContent = "Дата выезда должна быть позже даты заезда";
    return;
  }
  el.textContent = `${formatDate(s)} → ${formatDate(e)}, ${nightsText(n)}`;
}

startInput.addEventListener("change", () => { updateSummary(); renderCalendar(); });
endInput.addEventListener("change", () => { updateSummary(); renderCalendar(); });
roomInput.addEventListener("change", () => {
  const id = Number(roomInput.value);
  if (id > 0) selectRoom(id);
});

$("bookingForm").addEventListener("submit", async e => {
  e.preventDefault();
  const btn = $("submitBtn");
  const body = {
    room_id: Number(roomInput.value),
    start_at: startInput.value,
    end_at: endInput.value,
  };
  btn.disabled = true;
  try {
    const r = await call("POST", "/bookings", body);
    mark("POST /bookings", routeWorks(r));
    if (r.ok) {
      toast(`Готово! Бронь № ${r.data?.id ?? ""} создана`);
      pickingEnd = false;
      if (state.selectedRoom !== body.room_id) selectRoom(body.room_id);
      else loadBookings();
    } else {
      toast(r.data?.error || `Ошибка (код ${r.status})`, "err");
    }
  } catch {
    mark("POST /bookings", false);
    toast("Сервер не отвечает", "err");
  } finally {
    btn.disabled = false;
  }
});

// ================= Поиск свободных комнат =================

$("searchForm").addEventListener("submit", async e => {
  e.preventDefault();
  const s = $("searchStart").value, en = $("searchEnd").value;
  const box = $("searchResult");
  try {
    const r = await call("GET", `/rooms/available?start_at=${s}&end_at=${en}`);
    mark("GET /rooms/available", routeWorks(r));
    if (!r.ok) {
      box.innerHTML = r.status === 404 || r.status === 405
        ? `<div class="hint">Эндпоинт <code>GET /rooms/available</code> ещё не написан.</div>`
        : `<div class="hint">${escapeHtml(r.data?.error || "Ошибка " + r.status)}</div>`;
      return;
    }
    const rooms = Array.isArray(r.data) ? r.data : [];
    if (rooms.length === 0) {
      box.innerHTML = `<p class="empty">На эти даты всё занято.</p>`;
      return;
    }
    box.innerHTML = rooms.map(room =>
      `<button type="button" class="chip" data-id="${room.id}">${escapeHtml(room.name ?? "№ " + room.id)}</button>`
    ).join("");
    box.querySelectorAll(".chip").forEach(c => c.addEventListener("click", () => {
      startInput.value = s;
      endInput.value = en;
      updateSummary();
      selectRoom(Number(c.dataset.id));
      roomInput.scrollIntoView({ behavior: "smooth", block: "center" });
    }));
  } catch {
    mark("GET /rooms/available", false);
    toast("Сервер не отвечает", "err");
  }
});

// ================= Старт =================

(function init() {
  const t = today();
  startInput.value = toISO(addDays(t, 1));
  endInput.value = toISO(addDays(t, 3));
  $("searchStart").value = toISO(addDays(t, 1));
  $("searchEnd").value = toISO(addDays(t, 3));

  renderEndpoints();
  updateSummary();
  renderCalendar();
  checkServer();
  loadRooms();
})();
