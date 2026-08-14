"use strict";

// Data injected by the server as window.ROSEVIEW.
const data = window.ROSEVIEW;
const columns = data.Columns;
const rows = data.Rows;

const thead = document.getElementById("table-head");
const tbody = document.getElementById("table-body");
const search = document.getElementById("search");
const count = document.getElementById("count");

// Minimum query length before fuzzy matching kicks in. Shorter queries are
// ignored entirely so a stray single character does not flood the results.
const MIN_QUERY_LENGTH = 2;

// Per-column sort state: 0 = unsorted, 1 = ascending, 2 = descending.
const sortState = columns.map(() => 0);

// Fuse.js index over the rows, searched one word at a time (Fuse v6 has no
// built-in tokenized search). includeMatches yields the per-key character
// ranges used for highlighting. threshold 0.2 keeps the matching strict
// (about one wrong character per short word at most), and
// minMatchCharLength drops matches built on a single stray character.
const fuse = new Fuse(rows, {
  keys: columns.map((c) => c.field),
  threshold: 0.2,
  ignoreLocation: true,
  minMatchCharLength: 2,
  includeMatches: true,
});

const fieldToCol = new Map(columns.map((c, i) => [c.field, i]));

// searchRows splits the query into words and intersects the per-word hits,
// so a multi-word query can match across columns (e.g. "مبانی ریاضی" with
// the first word in the course name and the second in the faculty group).
// Returns a Map of refIndex -> per-column matched char indices.
function searchRows(query) {
  const words = query.split(/\s+/).filter((w) => w.length > 0);
  let acc = null;
  for (const word of words) {
    const hits = new Map();
    for (const result of fuse.search(word)) {
      const perCol = columns.map(() => new Set());
      for (const m of result.matches) {
        const c = fieldToCol.get(m.key);
        if (c === undefined) continue;
        for (const pair of m.indices) {
          for (let idx = pair[0]; idx <= pair[1]; idx++) perCol[c].add(idx);
        }
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
            for (const i of next[c]) perCol[c].add(i);
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

// renderCell emits a cell, wrapping matched runs in <mark>.
function renderCell(text, matched) {
  if (text === "") return emptyIcon();
  if (!matched || matched.length === 0) return escapeHtml(text);

  const ranges = [];
  let prev = -2;
  for (const idx of matched) {
    if (idx === prev + 1) ranges[ranges.length - 1].end = idx;
    else ranges.push({ start: idx, end: idx });
    prev = idx;
  }

  let html = "";
  let pos = 0;
  for (const r of ranges) {
    html += escapeHtml(text.slice(pos, r.start));
    html += "<mark>" + escapeHtml(text.slice(r.start, r.end + 1)) + "</mark>";
    pos = r.end + 1;
  }
  html += escapeHtml(text.slice(pos));
  return html;
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

// sortedIndexes returns row indexes ordered by the active sort column.
function sortedIndexes() {
  const idx = rows.map((_, i) => i);
  const col = sortState.findIndex((st) => st !== 0);
  const dir = sortState[col] === 1 ? 1 : -1;
  const field = columns[col].field;
  idx.sort((a, b) => compareCells(rows[a][field], rows[b][field]) * dir);
  return idx;
}

function renderHead() {
  let html = "<tr>";
  for (let i = 0; i < columns.length; i++) {
    const arrow =
      sortState[i] === 1
        ? '<svg class="sort-arrow-icon" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-sort-asc"></use></svg>'
        : sortState[i] === 2
          ? '<svg class="sort-arrow-icon" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><use href="#icon-sort-desc"></use></svg>'
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
  thead.innerHTML = html;
}

function renderRows() {
  const query = search.value.trim();
  const sorted = sortState.some((st) => st !== 0);

  // Queries shorter than the minimum are ignored entirely so a stray single
  // character does not flood the results with weak matches.
  const searching = query.length >= MIN_QUERY_LENGTH;

  // Run the query once against the Fuse index; rows are still iterated in
  // sorted order, filtered by the matched set.
  let matchSet = null;
  let matchMap = null;
  if (searching) {
    const results = searchRows(query);
    matchSet = new Set(results.keys());
    matchMap = new Map();
    for (const [ref, perCol] of results) {
      matchMap.set(
        ref,
        perCol.map((s) => [...s].sort((a, b) => a - b)),
      );
    }
  }

  let html = "";
  let shown = 0;
  const indexes = sorted ? sortedIndexes() : rows.map((_, i) => i);

  for (const i of indexes) {
    const row = rows[i];
    let matches = null;
    if (searching) {
      if (!matchSet.has(i)) continue;
      matches = matchMap.get(i);
    }

    shown++;
    html += "<tr>";
    for (let c = 0; c < columns.length; c++) {
      const value = row[columns[c].field] || "";
      html += "<td>" + renderCell(value, matches ? matches[c] : null) + "</td>";
    }
    html += "</tr>";
  }

  if (html === "") {
    html = '<tr class="empty-row"><td colspan="' + columns.length + '">نتیجهای یافت نشد</td></tr>';
  }
  tbody.innerHTML = html;
  count.textContent = shown + " نتیجه";
}

thead.addEventListener("click", (e) => {
  const th = e.target.closest("th");
  if (!th) return;

  const i = Number(th.dataset.col);
  sortState[i] = (sortState[i] + 1) % 3;
  for (let j = 0; j < sortState.length; j++) {
    if (j !== i) sortState[j] = 0;
  }

  renderHead();
  renderRows();
});

search.addEventListener("input", renderRows);

renderHead();
renderRows();

