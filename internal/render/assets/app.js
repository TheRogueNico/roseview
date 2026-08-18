"use strict";

// Data injected by the server as window.ROSEVIEW.
const data = window.ROSEVIEW;
const columns = data.Columns;
const rows = data.Rows;

const thead = document.getElementById("table-head");
const tbody = document.getElementById("table-body");
const pinnedWrap = document.getElementById("pinned-wrap");
const pinnedHead = document.getElementById("pinned-head");
const pinnedBody = document.getElementById("pinned-body");
const search = document.getElementById("search");
const count = document.getElementById("count");

// Per-column sort state: 0 = unsorted, 1 = ascending, 2 = descending. The
// main and pinned tables sort independently.
const sortState = columns.map(() => 0);
const pinnedSortState = columns.map(() => 0);

// Pinned rows, keyed by refIndex in insertion order. In-memory only.
const pinned = new Set();

// Fuse.js index over the rows, searched one word at a time (no
// built-in tokenized search). includeMatches yields the per-key character
// ranges used to decide which cells matched; threshold 0.3 keeps matching
// forgiving.
const fuse = new Fuse(rows, {
  keys: columns.map((c) => c.field),
  threshold: 0.3,
  ignoreLocation: true,
  includeMatches: true,
});

const fieldToCol = new Map(columns.map((c, i) => [c.field, i]));

// searchRows splits the query into words and intersects the per-word hits,
// so a multi-word query can match across columns (e.g. "مبانی ریاضی" with
// the first word in the course name and the second in the faculty group).
// Returns a Map of refIndex -> per-column matched flags.
function searchRows(query) {
  const words = query.split(/\s+/).filter((w) => w.length > 0);
  let acc = null;
  for (const word of words) {
    const hits = new Map();
    for (const result of fuse.search(word)) {
      const perCol = columns.map(() => false);
      for (const m of result.matches) {
        const c = fieldToCol.get(m.key);
        if (c === undefined) continue;
        if (m.indices.length > 0) perCol[c] = true;
      }
      hits.set(result.refIndex, perCol);
    }
    if (acc === null) {
      acc = hits;
    } else {
      for (const [ref, perCol] of acc) {
        const next = hits.get(ref);
        if (!next) {
          acc.delete(ref);
        } else {
          for (let c = 0; c < columns.length; c++) {
            perCol[c] = perCol[c] || next[c];
          }
        }
      }
    }
  }
  return acc;
}

function escapeHtml(s) {
  return s.replace(
    /[&<>"']/g,
    (c) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      })[c],
  );
}

// Convert Persian/Arabic digits to ASCII so numeric columns compare as
// numbers. Returns NaN when the string is not numeric.
function toNumber(s) {
  const digits =
    "\u06f0\u06f1\u06f2\u06f3\u06f4\u06f5\u06f6\u06f7\u06f8\u06f9\u0660\u0661\u0662\u0663\u0664\u0665\u0666\u0667\u0668\u0669";
  const t = s.replace(/[\u06f0-\u06f9\u0660-\u0669]/g, (d) => digits.indexOf(d) % 10);
  const n = Number(t);
  return Number.isNaN(n) ? NaN : n;
}

// renderCell emits a cell, wrapping the whole text in <mark> when the cell
// matched the query.
function renderCell(text, matched) {
  if (text === "") return emptyIcon();
  if (!matched) return escapeHtml(text);
  return "<mark>" + escapeHtml(text) + "</mark>";
}

function emptyIcon() {
  return '<svg class="empty-icon" width="14" height="14" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-minus"></use></svg>';
}

function compareCells(a, b) {
  const aEmpty = a === "" || a == null;
  const bEmpty = b === "" || b == null;
  if (aEmpty && bEmpty) return 0;
  if (aEmpty) return 1; // empties sort last
  if (bEmpty) return -1;

  const na = toNumber(a);
  const nb = toNumber(b);
  if (!Number.isNaN(na) && !Number.isNaN(nb)) return na - nb;
  return a.localeCompare(b, "fa");
}

// sortIndexes returns indexes ordered by the active column of state.
function sortIndexes(indexes, state) {
  const col = state.findIndex((st) => st !== 0);
  if (col < 0) return [...indexes];
  const dir = state[col] === 1 ? 1 : -1;
  const field = columns[col].field;
  return [...indexes].sort((a, b) => compareCells(rows[a][field], rows[b][field]) * dir);
}

// pinButton emits the bookmark button for row ref. Pinned rows show the
// minus icon so either table can unpin; everything else offers to pin.
function pinButton(ref) {
  const pinnedRow = pinned.has(ref);
  const icon = pinnedRow ? "minus" : "plus";
  const label = pinnedRow ? "حذف پین" : "پین کردن";
  return (
    '<button class="pin-btn' +
    (pinnedRow ? " pinned" : "") +
    '" data-pin="' +
    ref +
    '" aria-label="' +
    label +
    '"><svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-bookmark-' +
    icon +
    '"></use></svg></button>'
  );
}

// renderHead emits the column headers into theadEl. The pin column comes
// first; it has no data-col attribute and therefore cannot be sorted.
function renderHead(theadEl, state) {
  let html =
    '<tr><th class="pin-col"><svg class="pin-head-icon" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-bookmark"></use></svg></th>';
  for (let i = 0; i < columns.length; i++) {
    const arrow =
      state[i] === 1
        ? '<svg class="sort-arrow-icon" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-chevron-up"></use></svg>'
        : state[i] === 2
          ? '<svg class="sort-arrow-icon" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-chevron-down"></use></svg>'
          : "";
    html +=
      '<th data-col="' +
      i +
      '"><span>' +
      escapeHtml(columns[i].header) +
      '</span><span class="sort-arrow">' +
      arrow +
      "</span></th>";
  }
  html += "</tr>";
  theadEl.innerHTML = html;
}

// renderTable emits the given rows into tbodyEl, pin column first. A
// non-null matchMap filters rows to the matched set and highlights matched
// cells; null renders every row with no highlighting (the pinned table).
// The pinned table is hidden while empty, so the empty-row message only
// ever shows in the search results.
function renderTable(tbodyEl, indexes, matchMap) {
  const searching = matchMap !== null;
  let html = "";
  let shown = 0;
  for (const i of indexes) {
    const row = rows[i];
    let perCol = null;
    if (searching) {
      if (!matchMap.has(i)) continue;
      perCol = matchMap.get(i);
    }

    shown++;
    html += "<tr>";
    html += '<td class="pin-col">' + pinButton(i) + "</td>";
    for (let c = 0; c < columns.length; c++) {
      const value = row[columns[c].field] || "";
      html += "<td>" + renderCell(value, perCol ? perCol[c] : null) + "</td>";
    }
    html += "</tr>";
  }

  if (html === "") {
    html =
      '<tr class="empty-row"><td colspan="' +
      (columns.length + 1) +
      '">نتیجهای یافت نشد</td></tr>';
  }
  tbodyEl.innerHTML = html;
  return shown;
}

// renderMain renders the searchable results table.
function renderMain() {
  const query = search.value.trim();
  const sorted = sortState.some((st) => st !== 0);

  // An empty query shows all rows; anything else is handed to Fuse, which
  // decides match quality on its own.
  const searching = query.length > 0;

  // Run the query once against the Fuse index; rows are still iterated in
  // sorted order, filtered by the matched set.
  let matchMap = null;
  if (searching) {
    const results = searchRows(query);
    matchMap = new Map();
    for (const [ref, perCol] of results) {
      matchMap.set(ref, perCol);
    }
  }

  const indexes = sorted
    ? sortIndexes(rows.map((_, i) => i), sortState)
    : rows.map((_, i) => i);
  const shown = renderTable(tbody, indexes, matchMap);
  count.textContent = shown + " نتیجه";
}

// renderPinned renders the pinned table: search-independent, sorted by the
// pinned sort state, hidden while nothing is pinned.
function renderPinned() {
  const sorted = pinnedSortState.some((st) => st !== 0);
  const indexes = sorted ? sortIndexes([...pinned], pinnedSortState) : [...pinned];
  renderTable(pinnedBody, indexes, null);
  pinnedWrap.hidden = pinned.size === 0;
}

// cycleSort advances the sort state of a single column, clearing the rest.
function cycleSort(state, i) {
  state[i] = (state[i] + 1) % 3;
  for (let j = 0; j < state.length; j++) {
    if (j !== i) state[j] = 0;
  }
}

// togglePin adds or removes a row from the pinned set and refreshes both
// tables.
function togglePin(ref) {
  if (pinned.has(ref)) {
    pinned.delete(ref);
  } else {
    pinned.add(ref);
  }
  renderMain();
  renderPinned();
}

thead.addEventListener("click", (e) => {
  const th = e.target.closest("th");
  if (!th || th.dataset.col === undefined) return;

  cycleSort(sortState, Number(th.dataset.col));
  renderHead(thead, sortState);
  renderMain();
});

pinnedHead.addEventListener("click", (e) => {
  const th = e.target.closest("th");
  if (!th || th.dataset.col === undefined) return;

  cycleSort(pinnedSortState, Number(th.dataset.col));
  renderHead(pinnedHead, pinnedSortState);
  renderPinned();
});

tbody.addEventListener("click", (e) => {
  const btn = e.target.closest(".pin-btn");
  if (!btn) return;
  togglePin(Number(btn.dataset.pin));
});

pinnedBody.addEventListener("click", (e) => {
  const btn = e.target.closest(".pin-btn");
  if (!btn) return;
  togglePin(Number(btn.dataset.pin));
});

// Typing is debounced so a burst of keystrokes does not rebuild the whole
// table once per character; only the settled query is rendered.
let searchTimer = null;
search.addEventListener("input", () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(renderMain, 120);
});

renderHead(thead, sortState);
renderHead(pinnedHead, pinnedSortState);
renderMain();
renderPinned();