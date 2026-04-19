(function () {
  'use strict';

  // ── Piece images ──────────────────────────────────────────────────────────
  const PIECE_CODES = ['wK','wQ','wR','wB','wN','wP','bK','bQ','bR','bB','bN','bP'];
  const pieceImages = {};
  let piecesReady = null;

  function loadPieces() {
    if (piecesReady) return piecesReady;
    piecesReady = Promise.all(PIECE_CODES.map(code => new Promise(res => {
      const img = new Image();
      img.onload = () => { pieceImages[code] = img; res(); };
      img.onerror = res;
      img.src = '/static/pieces/' + code + '.svg';
    })));
    return piecesReady;
  }

  // ── FEN ───────────────────────────────────────────────────────────────────
  function parseFEN(fen) {
    const board = Array.from({length: 8}, () => new Array(8).fill(null));
    if (!fen) return board;
    const rows = fen.split(' ')[0].split('/');
    if (rows.length !== 8) return board;
    for (let i = 0; i < 8; i++) {
      let file = 0;
      for (const c of rows[i]) {
        if (c >= '1' && c <= '8') { file += c.charCodeAt(0) - 48; continue; }
        const color = (c >= 'A' && c <= 'Z') ? 'w' : 'b';
        const rank = 7 - i;
        if (file < 8) board[rank][file] = color + c.toUpperCase();
        file++;
      }
    }
    return board;
  }

  // ── Theme ─────────────────────────────────────────────────────────────────
  const C = {
    bg: '#fafafa', headerBg: '#222', headerText: '#fff',
    text: '#222', muted: '#777', border: '#ddd',
    btnBg: '#fff', btnBorder: '#bbb', btnHover: '#eee',
    btnPrimBg: '#456', btnPrimBorder: '#345', btnPrimText: '#fff',
    tableHead: '#f3f3f3', rowAlt: '#f5f5f5',
    active: '#f6ff99', flash: '#e6f4e6', flashText: '#244',
    errorBg: '#fbe3e3', errorText: '#611',
    link: '#0055cc',
    lightSq: '#f0d9b5', darkSq: '#b58863', boardBorder: '#8b6d4b',
    scrollbar: '#bbb',
  };

  const PAD = 14;
  const HEADER_H = 48;

  // ── App state ─────────────────────────────────────────────────────────────
  const S = {
    pd: null, canvas: null, ctx: null,
    W: 0, H: 0, dpr: 1,
    // game
    ply: 0, fen: '',
    movesScroll: 0, movesContentH: 0,
    showFen: false,
    bookmarkNote: '',
    // index
    gamesScroll: 0, gamesContentH: 0,
    // import
    pgnText: '',
    // input overlay
    ovrActive: false, ovrId: null, ovrCommit: null,
    pgnActive: false,
    // hover
    hoverId: null,
    // touch scroll
    _tY0: 0, _tS0: 0,
  };

  let HITS = [];

  function hit(id, x, y, w, h, fn) {
    HITS.push({id, x, y, w, h, fn});
  }

  function hitAt(cx, cy) {
    for (let i = HITS.length - 1; i >= 0; i--) {
      const h = HITS[i];
      if (cx >= h.x && cx < h.x + h.w && cy >= h.y && cy < h.y + h.h) return h;
    }
    return null;
  }

  // ── Draw primitives ───────────────────────────────────────────────────────
  function rrect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.arcTo(x + w, y, x + w, y + r, r);
    ctx.lineTo(x + w, y + h - r);
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r);
    ctx.lineTo(x + r, y + h);
    ctx.arcTo(x, y + h, x, y + h - r, r);
    ctx.lineTo(x, y + r);
    ctx.arcTo(x, y, x + r, y, r);
    ctx.closePath();
  }

  function clip(ctx, text, maxW) {
    if (ctx.measureText(text).width <= maxW) return text;
    let t = text;
    while (t.length > 1 && ctx.measureText(t + '…').width > maxW) t = t.slice(0, -1);
    return t + '…';
  }

  function btn(ctx, id, label, x, y, w, h, primary) {
    const hov = S.hoverId === id;
    ctx.fillStyle = primary ? C.btnPrimBg : (hov ? C.btnHover : C.btnBg);
    rrect(ctx, x, y, w, h, 4);
    ctx.fill();
    ctx.strokeStyle = primary ? C.btnPrimBorder : C.btnBorder;
    ctx.lineWidth = 1;
    ctx.stroke();
    ctx.fillStyle = primary ? C.btnPrimText : C.text;
    ctx.font = '13px system-ui,sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(label, x + w / 2, y + h / 2);
  }

  function inputBox(ctx, x, y, w, h, value, placeholder) {
    ctx.fillStyle = '#fff';
    rrect(ctx, x, y, w, h, 4);
    ctx.fill();
    ctx.strokeStyle = C.border;
    ctx.lineWidth = 1;
    ctx.stroke();
    ctx.font = '13px system-ui,sans-serif';
    ctx.textBaseline = 'middle';
    ctx.textAlign = 'left';
    if (value) {
      ctx.fillStyle = C.text;
      ctx.fillText(clip(ctx, value, w - 14), x + 7, y + h / 2);
    } else {
      ctx.fillStyle = C.muted;
      ctx.fillText(clip(ctx, placeholder, w - 14), x + 7, y + h / 2);
    }
  }

  // ── Header ────────────────────────────────────────────────────────────────
  function drawHeader(ctx, pd) {
    ctx.fillStyle = C.headerBg;
    ctx.fillRect(0, 0, S.W, HEADER_H);
    // Title
    ctx.fillStyle = C.headerText;
    ctx.font = 'bold 15px system-ui,sans-serif';
    ctx.textAlign = 'left';
    ctx.textBaseline = 'middle';
    ctx.fillText('chess_notation', PAD, HEADER_H / 2);
    hit('home', 0, 0, 170, HEADER_H, () => nav('/'));

    if (pd.page === 'index') {
      const bw = 108, bh = 30, bx = S.W - PAD - bw, by = (HEADER_H - bh) / 2;
      btn(ctx, 'import-nav', '+ Import PGN', bx, by, bw, bh, false);
      hit('import-nav', bx, by, bw, bh, () => nav('/import'));
    } else {
      const bw = 84, bh = 30, bx = S.W - PAD - bw, by = (HEADER_H - bh) / 2;
      btn(ctx, 'back', '← Games', bx, by, bw, bh, false);
      hit('back', bx, by, bw, bh, () => nav('/'));
    }
  }

  // ── Chess board ───────────────────────────────────────────────────────────
  function drawBoard(ctx, fen, bx, by, bsize) {
    const sq = bsize / 8;
    const board = parseFEN(fen);

    // Squares
    for (let r = 7; r >= 0; r--) {
      for (let f = 0; f < 8; f++) {
        const x = bx + f * sq, y = by + (7 - r) * sq;
        ctx.fillStyle = (f + r) % 2 === 0 ? C.darkSq : C.lightSq;
        ctx.fillRect(x, y, sq, sq);
      }
    }
    // Pieces
    for (let r = 7; r >= 0; r--) {
      for (let f = 0; f < 8; f++) {
        const piece = board[r][f];
        if (!piece) continue;
        const img = pieceImages[piece];
        if (!img || !img.complete) continue;
        const x = bx + f * sq, y = by + (7 - r) * sq;
        const pad = sq * 0.06;
        ctx.drawImage(img, x + pad, y + pad, sq - 2 * pad, sq - 2 * pad);
      }
    }
    // Coordinates
    const fs = Math.max(9, Math.round(sq * 0.19));
    ctx.font = fs + 'px system-ui,sans-serif';
    for (let f = 0; f < 8; f++) {
      const x = bx + f * sq;
      ctx.fillStyle = f % 2 === 0 ? C.lightSq : C.darkSq;
      ctx.textBaseline = 'bottom'; ctx.textAlign = 'left';
      ctx.fillText(String.fromCharCode(97 + f), x + sq * 0.06, by + bsize - sq * 0.06);
    }
    for (let r = 0; r < 8; r++) {
      const y = by + (7 - r) * sq;
      ctx.fillStyle = (7 + r) % 2 === 0 ? C.lightSq : C.darkSq;
      ctx.textBaseline = 'top'; ctx.textAlign = 'right';
      ctx.fillText(String(r + 1), bx + bsize - sq * 0.06, y + sq * 0.06);
    }
  }

  // ── Game page ─────────────────────────────────────────────────────────────
  function drawGamePage(ctx, pd) {
    const narrow = S.W < 520;
    const cy = HEADER_H;
    const ch = S.H - cy;

    // Game meta header
    ctx.fillStyle = C.text;
    ctx.font = 'bold 14px system-ui,sans-serif';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.fillText(clip(ctx, pd.white + ' vs ' + pd.black, S.W - 2 * PAD), PAD, cy + 18);

    ctx.fillStyle = C.muted;
    ctx.font = '11px system-ui,sans-serif';
    const meta = [pd.event, pd.site, pd.date, 'Result: ' + pd.result].filter(Boolean).join(' — ');
    ctx.fillText(clip(ctx, meta, S.W - 2 * PAD), PAD, cy + 36);

    const bodyY = cy + 50;
    const bodyH = ch - 50;

    let bx, by, bsize, sideX, sideY, sideW, sideH;

    if (!narrow) {
      // Side by side: board left, controls right
      bsize = Math.max(180, Math.min(bodyH - PAD * 2, (S.W - PAD * 3) * 0.52));
      bsize = Math.floor(bsize / 8) * 8;
      bx = PAD; by = bodyY + (bodyH - bsize) / 2;
      sideX = bx + bsize + PAD * 2;
      sideY = bodyY;
      sideW = S.W - sideX - PAD;
      sideH = bodyH;
    } else {
      // Stacked: board top, controls bottom
      bsize = Math.max(160, Math.min(S.W - PAD * 2, (bodyH - PAD * 2) * 0.52));
      bsize = Math.floor(bsize / 8) * 8;
      bx = Math.floor((S.W - bsize) / 2); by = bodyY + PAD;
      sideX = PAD; sideY = by + bsize + PAD;
      sideW = S.W - PAD * 2; sideH = S.H - sideY - PAD;
    }

    // Board border + board
    ctx.strokeStyle = C.boardBorder;
    ctx.lineWidth = 1;
    ctx.strokeRect(bx - 1, by - 1, bsize + 2, bsize + 2);
    drawBoard(ctx, S.fen, bx, by, bsize);

    // Nav buttons
    const navH = 40, btnW = 44, btnH = 36, gap = 5;
    const navBtns = [
      {id: 'nav-first', lbl: '⏮', ply: 0},
      {id: 'nav-prev',  lbl: '◀',  ply: Math.max(0, S.ply - 1)},
      {id: 'nav-next',  lbl: '▶',  ply: Math.min(pd.max_ply, S.ply + 1)},
      {id: 'nav-last',  lbl: '⏭', ply: pd.max_ply},
    ];
    let nx = sideX;
    const ny = sideY + 4;
    for (const b of navBtns) {
      btn(ctx, b.id, b.lbl, nx, ny, btnW, btnH, false);
      const tp = b.ply;
      hit(b.id, nx, ny, btnW, btnH, () => gotoPlay(pd, tp));
      nx += btnW + gap;
    }
    ctx.fillStyle = C.muted;
    ctx.font = '13px system-ui,sans-serif';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.fillText(S.ply + ' / ' + pd.max_ply, nx + 6, ny + btnH / 2);

    // Move list
    const mlY = sideY + navH + 8;
    let mlH = sideH - navH - 8;

    // Bookmark section at bottom
    const bmH = 40;
    const bookmarksH = pd.bookmarks && pd.bookmarks.length > 0 ? Math.min(pd.bookmarks.length * 22 + 8, 80) : 0;
    const fenH = S.showFen ? 36 : 0;
    const sectionH = bmH + bookmarksH + fenH + 28;
    mlH = Math.max(60, sideH - navH - 8 - sectionH);

    drawMoveList(ctx, pd, sideX, mlY, sideW, mlH);

    // Bookmark form
    const bfY = mlY + mlH + 6;
    drawBookmarkSection(ctx, pd, sideX, bfY, sideW);
  }

  function drawMoveList(ctx, pd, x, y, w, h) {
    ctx.save();
    ctx.beginPath(); ctx.rect(x, y, w, h); ctx.clip();

    const itemH = 26;
    const colW = Math.floor(w / 2);
    const moves = pd.moves || [];
    const totalH = Math.ceil(moves.length / 2) * itemH;
    S.movesContentH = totalH;

    // Clamp scroll
    S.movesScroll = Math.max(0, Math.min(Math.max(0, totalH - h), S.movesScroll));

    ctx.fillStyle = C.bg;
    ctx.fillRect(x, y, w, h);
    ctx.strokeStyle = C.border;
    ctx.lineWidth = 0.5;
    ctx.strokeRect(x, y, w, h);

    for (let i = 0; i < moves.length; i++) {
      const m = moves[i];
      const col = i % 2, row = Math.floor(i / 2);
      const mx = x + col * colW;
      const my = y + row * itemH - S.movesScroll;
      if (my + itemH < y || my > y + h) continue;

      if (m.ply === S.ply) {
        ctx.fillStyle = C.active;
        ctx.fillRect(mx, my, colW, itemH);
      } else if (row % 2 === 0 && col === 0) {
        ctx.fillStyle = C.rowAlt;
        ctx.fillRect(mx, my, colW * 2, itemH);
      }

      ctx.textBaseline = 'middle'; ctx.textAlign = 'left';
      if (col === 0) {
        ctx.fillStyle = C.muted;
        ctx.font = '11px system-ui,sans-serif';
        ctx.fillText(m.number + '.', mx + 4, my + itemH / 2);
      }
      ctx.fillStyle = C.link;
      ctx.font = '13px system-ui,sans-serif';
      ctx.fillText(clip(ctx, m.san, colW - (col === 0 ? 30 : 8)), mx + (col === 0 ? 28 : 6), my + itemH / 2);

      hit('mv-' + m.ply, mx, my, colW, itemH, () => gotoPlay(pd, m.ply));
    }

    // Scrollbar
    if (totalH > h) {
      const sbW = 4, sbH = Math.max(24, (h / totalH) * h);
      const sbY = y + (S.movesScroll / Math.max(1, totalH - h)) * (h - sbH);
      ctx.fillStyle = C.scrollbar;
      ctx.fillRect(x + w - sbW - 1, sbY, sbW, sbH);
    }
    ctx.restore();
  }

  function drawBookmarkSection(ctx, pd, x, y, w) {
    const inputW = w - 108, inputH = 30, btnW = 96, gap = 6;

    // Note input
    inputBox(ctx, x, y, inputW, inputH, S.bookmarkNote, 'Bookmark note…');
    hit('bm-input', x, y, inputW, inputH, () => {
      activateOvr('bm-note', x, y, inputW, inputH, S.bookmarkNote, val => { S.bookmarkNote = val; });
    });

    // Bookmark button
    btn(ctx, 'bm-btn', 'Bookmark ply ' + S.ply, x + inputW + gap, y, btnW, inputH, false);
    hit('bm-btn', x + inputW + gap, y, btnW, inputH, () => submitBookmark(pd));

    let cy2 = y + inputH + 6;

    // Existing bookmarks
    if (pd.bookmarks && pd.bookmarks.length) {
      ctx.fillStyle = C.muted;
      ctx.font = '11px system-ui,sans-serif';
      ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
      for (const bm of pd.bookmarks) {
        const label = 'ply ' + bm.ply + (bm.note ? ' — ' + bm.note : '');
        ctx.fillStyle = C.link;
        ctx.fillText(clip(ctx, label, w - 4), x + 2, cy2 + 11);
        hit('bm-goto-' + bm.ply, x, cy2, w, 22, () => gotoPlay(pd, bm.ply));
        cy2 += 22;
      }
    }

    // FEN toggle
    const fenLabel = S.showFen ? '▲ FEN' : '▼ FEN';
    ctx.fillStyle = C.muted;
    ctx.font = '11px system-ui,sans-serif';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.fillText(fenLabel, x, cy2 + 10);
    hit('fen-toggle', x, cy2, 60, 20, () => { S.showFen = !S.showFen; render(); });
    cy2 += 22;

    if (S.showFen && S.fen) {
      ctx.fillStyle = '#eee';
      ctx.fillRect(x, cy2, w, 24);
      ctx.strokeStyle = C.border; ctx.lineWidth = 1;
      ctx.strokeRect(x, cy2, w, 24);
      ctx.fillStyle = C.text;
      ctx.font = '10px ui-monospace,monospace';
      ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
      ctx.fillText(clip(ctx, S.fen, w - 8), x + 4, cy2 + 12);
    }
  }

  async function gotoPlay(pd, n) {
    n = Math.max(0, Math.min(pd.max_ply, n));
    if (n === S.ply) return;
    const r = await fetch('/game/' + pd.game_id + '/fragment/' + n, {headers: {'Accept': 'text/plain'}});
    if (!r.ok) return;
    S.fen = (await r.text()).trim();
    S.ply = n;
    // Scroll move list to keep active move visible
    const moves = pd.moves || [];
    const idx = moves.findIndex(m => m.ply === n);
    if (idx >= 0) {
      const itemH = 26;
      const row = Math.floor(idx / 2);
      const itemY = row * itemH;
      const listH = S.movesContentH; // approximate
      if (itemY < S.movesScroll) S.movesScroll = itemY;
      if (itemY + itemH > S.movesScroll + 200) S.movesScroll = itemY + itemH - 200;
    }
    history.replaceState(null, '', '/game/' + pd.game_id + '/ply/' + n);
    render();
  }

  function submitBookmark(pd) {
    const form = document.getElementById('bookmark-form');
    if (!form) return;
    document.getElementById('bookmark-ply').value = S.ply;
    document.getElementById('bookmark-note-val').value = S.bookmarkNote;
    form.action = '/game/' + pd.game_id + '/bookmark';
    form.submit();
  }

  // ── Index page ────────────────────────────────────────────────────────────
  // Filter state is kept in pd.filter directly (mutated locally)
  function drawIndexPage(ctx, pd) {
    let y = HEADER_H + PAD;

    // Flash
    if (pd.imported > 0) {
      const fh = 32;
      ctx.fillStyle = C.flash;
      rrect(ctx, PAD, y, S.W - 2*PAD, fh, 4);
      ctx.fill();
      ctx.fillStyle = C.flashText;
      ctx.font = '13px system-ui,sans-serif';
      ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
      ctx.fillText('Imported ' + pd.imported + ' game' + (pd.imported !== 1 ? 's' : '') + '.', PAD + 8, y + fh / 2);
      y += fh + 8;
    }

    // Title
    ctx.fillStyle = C.text;
    ctx.font = 'bold 16px system-ui,sans-serif';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.fillText('Games (' + pd.total + ')', PAD, y + 14);
    y += 32;

    // Filter row
    y = drawFilterRow(ctx, pd, y);
    y += 8;

    // Games table
    const tableY = y;
    const tableH = S.H - tableY - 48;
    drawGamesTable(ctx, pd, PAD, tableY, S.W - 2*PAD, tableH);

    // Pagination
    drawPagination(ctx, pd, PAD, S.H - 44, S.W - 2*PAD, 40);
  }

  function drawFilterRow(ctx, pd, y) {
    const narrow = S.W < 480;
    const fh = narrow ? 50 : 52;
    ctx.fillStyle = '#f3f3f3';
    rrect(ctx, PAD, y, S.W - 2*PAD, fh, 4);
    ctx.fill();
    ctx.strokeStyle = C.border; ctx.lineWidth = 1; ctx.stroke();

    const iy = y + (fh - 30) / 2;

    if (narrow) {
      // Compact: single filter button + clear
      const cur = [pd.filter.white, pd.filter.black, pd.filter.result].filter(Boolean).join(', ');
      ctx.fillStyle = C.muted; ctx.font = '12px system-ui,sans-serif';
      ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
      ctx.fillText(cur ? 'Filter: ' + clip(ctx, cur, S.W - 2*PAD - 100) : 'No filter', PAD + 8, y + fh / 2);
      const bw = 60;
      btn(ctx, 'filter-clear', 'Clear', S.W - PAD - bw, iy, bw, 30, false);
      hit('filter-clear', S.W - PAD - bw, iy, bw, 30, () => nav('/'));
      return y + fh;
    }

    let ix = PAD + 8;
    const iw = 96, ih = 30, gap = 8;

    ctx.fillStyle = C.muted; ctx.font = '10px system-ui,sans-serif';
    ctx.textBaseline = 'top'; ctx.textAlign = 'left';

    // White
    ctx.fillText('White', ix, y + 7);
    inputBox(ctx, ix, iy, iw, ih, pd.filter.white, 'name');
    hit('f-white', ix, iy, iw, ih, () => activateOvr('f-white', ix, iy, iw, ih, pd.filter.white || '', val => { pd.filter.white = val; render(); }));
    ix += iw + gap;

    // Black
    ctx.fillText('Black', ix, y + 7);
    inputBox(ctx, ix, iy, iw, ih, pd.filter.black, 'name');
    hit('f-black', ix, iy, iw, ih, () => activateOvr('f-black', ix, iy, iw, ih, pd.filter.black || '', val => { pd.filter.black = val; render(); }));
    ix += iw + gap;

    // Result (cycle on click)
    const RESULTS = ['', '1-0', '0-1', '1/2-1/2', '*'];
    ctx.fillText('Result', ix, y + 7);
    inputBox(ctx, ix, iy, 78, ih, pd.filter.result, '(any)');
    hit('f-result', ix, iy, 78, ih, () => {
      const idx = RESULTS.indexOf(pd.filter.result);
      pd.filter.result = RESULTS[(idx + 1) % RESULTS.length];
      render();
    });
    ix += 78 + gap;

    // Filter & Clear
    btn(ctx, 'f-apply', 'Filter', ix, iy, 60, ih, false);
    hit('f-apply', ix, iy, 60, ih, () => applyFilter(pd));
    ix += 66;

    btn(ctx, 'f-clear', 'Clear', ix, iy, 52, ih, false);
    hit('f-clear', ix, iy, 52, ih, () => nav('/'));

    return y + fh;
  }

  function applyFilter(pd) {
    const p = new URLSearchParams();
    if (pd.filter.white) p.set('white', pd.filter.white);
    if (pd.filter.black) p.set('black', pd.filter.black);
    if (pd.filter.result) p.set('result', pd.filter.result);
    if (pd.per_page && pd.per_page !== 50) p.set('per', pd.per_page);
    nav('/?' + p.toString());
  }

  function drawGamesTable(ctx, pd, x, y, w, h) {
    ctx.save();
    ctx.beginPath(); ctx.rect(x, y, w, h); ctx.clip();

    const narrow = S.W < 480;
    const rowH = 36, headH = 34;
    const games = pd.games || [];
    S.gamesContentH = games.length * rowH;
    S.gamesScroll = Math.max(0, Math.min(Math.max(0, S.gamesContentH - (h - headH)), S.gamesScroll));

    // Column definitions
    let cols;
    if (narrow) {
      const rW = w - 36 - 60; // leftover for White+Black
      cols = [
        {lbl:'#',      w: 36},
        {lbl:'White',  w: Math.floor(rW * 0.5)},
        {lbl:'Black',  w: Math.floor(rW * 0.5)},
        {lbl:'Result', w: 60},
      ];
    } else {
      const rW = w - 40 - 70 - 90; // leftover for White+Black+Event
      cols = [
        {lbl:'#',      w: 40},
        {lbl:'White',  w: Math.floor(rW * 0.3)},
        {lbl:'Black',  w: Math.floor(rW * 0.3)},
        {lbl:'Event',  w: Math.floor(rW * 0.4)},
        {lbl:'Date',   w: 90},
        {lbl:'Result', w: 70},
      ];
    }
    // Compute x offsets
    const cxs = [x];
    for (let i = 0; i < cols.length - 1; i++) cxs.push(cxs[i] + cols[i].w);

    // Header
    ctx.fillStyle = C.tableHead;
    ctx.fillRect(x, y, w, headH);
    ctx.strokeStyle = C.border; ctx.lineWidth = 1;
    ctx.strokeRect(x, y, w, headH);
    ctx.fillStyle = C.text;
    ctx.font = 'bold 11px system-ui,sans-serif';
    ctx.textBaseline = 'middle'; ctx.textAlign = 'left';
    for (let i = 0; i < cols.length; i++) {
      ctx.fillText(cols[i].lbl, cxs[i] + 6, y + headH / 2);
    }

    // Rows
    const dataY = y + headH;
    for (let gi = 0; gi < games.length; gi++) {
      const g = games[gi];
      const ry = dataY + gi * rowH - S.gamesScroll;
      if (ry + rowH < y || ry > y + h) continue;

      ctx.fillStyle = gi % 2 === 0 ? C.bg : C.rowAlt;
      ctx.fillRect(x, ry, w, rowH);
      ctx.strokeStyle = C.border; ctx.lineWidth = 0.5;
      ctx.beginPath(); ctx.moveTo(x, ry + rowH); ctx.lineTo(x + w, ry + rowH); ctx.stroke();

      const vals = narrow
        ? [String(g.id), g.white, g.black, g.result]
        : [String(g.id), g.white, g.black, g.event || '', g.date || '', g.result];

      ctx.textBaseline = 'middle'; ctx.textAlign = 'left';
      for (let ci = 0; ci < cols.length; ci++) {
        ctx.fillStyle = ci === 1 ? C.link : C.text;
        ctx.font = ci === 1 ? 'bold 12px system-ui,sans-serif' : '12px system-ui,sans-serif';
        ctx.fillText(clip(ctx, vals[ci] || '', cols[ci].w - 12), cxs[ci] + 6, ry + rowH / 2);
      }
      hit('game-' + g.id, x, ry, w, rowH, () => nav('/game/' + g.id));
    }

    // Scrollbar
    if (S.gamesContentH > h - headH) {
      const avail = h - headH;
      const sbH = Math.max(24, (avail / S.gamesContentH) * avail);
      const sbY = dataY + (S.gamesScroll / Math.max(1, S.gamesContentH - avail)) * (avail - sbH);
      ctx.fillStyle = C.scrollbar;
      ctx.fillRect(x + w - 4, sbY, 4, sbH);
    }

    ctx.restore();
  }

  function drawPagination(ctx, pd, x, y, w, h) {
    const btnW = 76, btnH = 32, by2 = y + (h - btnH) / 2;
    const cx = x + w / 2;

    if (pd.has_prev) {
      btn(ctx, 'pg-prev', '◀ Prev', cx - btnW - 60, by2, btnW, btnH, false);
      const pp = pd.prev_page;
      hit('pg-prev', cx - btnW - 60, by2, btnW, btnH, () => navPage(pd, pp));
    } else {
      ctx.globalAlpha = 0.35;
      btn(ctx, 'pg-prev-d', '◀ Prev', cx - btnW - 60, by2, btnW, btnH, false);
      ctx.globalAlpha = 1;
    }

    ctx.fillStyle = C.text; ctx.font = '13px system-ui,sans-serif';
    ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    ctx.fillText('Page ' + pd.page_num + ' / ' + pd.total_pages, cx, y + h / 2);

    if (pd.has_next) {
      btn(ctx, 'pg-next', 'Next ▶', cx + 60, by2, btnW, btnH, false);
      const np = pd.next_page;
      hit('pg-next', cx + 60, by2, btnW, btnH, () => navPage(pd, np));
    } else {
      ctx.globalAlpha = 0.35;
      btn(ctx, 'pg-next-d', 'Next ▶', cx + 60, by2, btnW, btnH, false);
      ctx.globalAlpha = 1;
    }
  }

  function navPage(pd, page) {
    const p = new URLSearchParams();
    if (pd.filter.white) p.set('white', pd.filter.white);
    if (pd.filter.black) p.set('black', pd.filter.black);
    if (pd.filter.result) p.set('result', pd.filter.result);
    if (pd.per_page && pd.per_page !== 50) p.set('per', pd.per_page);
    p.set('page', page);
    nav('/?' + p.toString());
  }

  // ── Import page ───────────────────────────────────────────────────────────
  function drawImportPage(ctx, pd) {
    let y = HEADER_H + PAD;

    if (pd.error) {
      const eh = 38;
      ctx.fillStyle = C.errorBg;
      rrect(ctx, PAD, y, S.W - 2*PAD, eh, 4);
      ctx.fill();
      ctx.fillStyle = C.errorText; ctx.font = '13px system-ui,sans-serif';
      ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
      ctx.fillText(clip(ctx, pd.error, S.W - 2*PAD - 16), PAD + 8, y + eh / 2);
      y += eh + 10;
    }

    ctx.fillStyle = C.text; ctx.font = 'bold 17px system-ui,sans-serif';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.fillText('Import PGN', PAD, y + 14);
    y += 36;

    ctx.fillStyle = C.muted; ctx.font = '13px system-ui,sans-serif';
    ctx.fillText('Upload a .pgn file or paste PGN text below.', PAD, y + 10);
    y += 30;

    // File button
    const fbW = 160, fbH = 36;
    btn(ctx, 'file-btn', '📂 Choose .pgn file', PAD, y, fbW, fbH, false);
    hit('file-btn', PAD, y, fbW, fbH, triggerFile);
    y += fbH + 10;

    // Divider
    ctx.fillStyle = C.muted; ctx.font = 'italic 13px system-ui,sans-serif';
    ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    ctx.fillText('— or paste PGN text —', S.W / 2, y + 9);
    y += 26;

    // PGN textarea visual
    const taW = S.W - 2*PAD;
    const taH = Math.max(100, Math.min(320, S.H - y - 70));
    ctx.fillStyle = '#fff';
    rrect(ctx, PAD, y, taW, taH, 4);
    ctx.fill();
    ctx.strokeStyle = C.border; ctx.lineWidth = 1; ctx.stroke();

    ctx.save();
    ctx.beginPath(); ctx.rect(PAD + 4, y + 4, taW - 8, taH - 8); ctx.clip();
    if (S.pgnText) {
      ctx.fillStyle = C.text; ctx.font = '12px ui-monospace,monospace';
      ctx.textAlign = 'left'; ctx.textBaseline = 'top';
      const lines = S.pgnText.split('\n');
      for (let i = 0; i < Math.min(lines.length, Math.floor((taH - 8) / 16)); i++) {
        ctx.fillText(lines[i], PAD + 8, y + 8 + i * 16);
      }
    } else {
      ctx.fillStyle = C.muted; ctx.font = '12px ui-monospace,monospace';
      ctx.textAlign = 'left'; ctx.textBaseline = 'top';
      ['[Event "?"]', '[White "..."]', '[Black "..."]', '', '1. e4 e5 ...'].forEach((l, i) => {
        ctx.fillText(l, PAD + 8, y + 8 + i * 16);
      });
    }
    ctx.restore();

    hit('pgn-area', PAD, y, taW, taH, () => activatePgn(PAD, y, taW, taH));
    y += taH + 12;

    // Action buttons
    const abW = 100, abH = 36;
    btn(ctx, 'imp-submit', 'Import', PAD, y, abW, abH, true);
    hit('imp-submit', PAD, y, abW, abH, doImport);

    btn(ctx, 'imp-cancel', 'Cancel', PAD + abW + 10, y, 80, abH, false);
    hit('imp-cancel', PAD + abW + 10, y, 80, abH, () => nav('/'));
  }

  function triggerFile() {
    const fi = document.getElementById('import-file');
    if (fi) fi.click();
  }

  function activatePgn(x, y, w, h) {
    const ta = document.getElementById('ovr-pgn');
    ta.value = S.pgnText;
    Object.assign(ta.style, {
      left: x + 'px', top: y + 'px', width: w + 'px', height: h + 'px',
      opacity: '1', pointerEvents: 'all', zIndex: '20',
    });
    ta.focus();
    S.pgnActive = true;
  }

  function deactivatePgn() {
    const ta = document.getElementById('ovr-pgn');
    S.pgnText = ta.value;
    Object.assign(ta.style, {opacity: '0', pointerEvents: 'none'});
    S.pgnActive = false;
    render();
  }

  function doImport() {
    const ta = document.getElementById('ovr-pgn');
    if (S.pgnActive) S.pgnText = ta.value;
    document.getElementById('import-pgn-text').value = S.pgnText;
    document.getElementById('import-form').submit();
  }

  // ── Overlay input ─────────────────────────────────────────────────────────
  function activateOvr(id, x, y, w, h, value, commit) {
    const inp = document.getElementById('ovr-input');
    inp.value = value;
    Object.assign(inp.style, {
      left: x + 'px', top: y + 'px', width: w + 'px', height: h + 'px',
      opacity: '1', pointerEvents: 'all', zIndex: '20',
    });
    inp.focus();
    S.ovrActive = true;
    S.ovrId = id;
    S.ovrCommit = commit;
  }

  function deactivateOvr() {
    if (!S.ovrActive) return;
    const inp = document.getElementById('ovr-input');
    if (S.ovrCommit) S.ovrCommit(inp.value);
    Object.assign(inp.style, {opacity: '0', pointerEvents: 'none'});
    S.ovrActive = false; S.ovrId = null; S.ovrCommit = null;
    render();
  }

  // ── Render ────────────────────────────────────────────────────────────────
  function render() {
    const pd = S.pd;
    if (!pd || !S.ctx) return;
    const ctx = S.ctx;

    ctx.save();
    ctx.scale(S.dpr, S.dpr);
    ctx.clearRect(0, 0, S.W, S.H);
    HITS = [];

    ctx.fillStyle = C.bg;
    ctx.fillRect(0, 0, S.W, S.H);

    drawHeader(ctx, pd);

    switch (pd.page) {
      case 'index':  drawIndexPage(ctx, pd);  break;
      case 'game':   drawGamePage(ctx, pd);   break;
      case 'import': drawImportPage(ctx, pd); break;
    }

    ctx.restore();
  }

  // ── Navigation ────────────────────────────────────────────────────────────
  function nav(url) { window.location.href = url; }

  // ── Events ────────────────────────────────────────────────────────────────
  function canvasPos(e) {
    const r = S.canvas.getBoundingClientRect();
    const src = e.touches ? e.touches[0] : (e.changedTouches ? e.changedTouches[0] : e);
    return {x: src.clientX - r.left, y: src.clientY - r.top};
  }

  function onClickCanvas(e) {
    const {x, y} = canvasPos(e);
    const h = hitAt(x, y);
    if (h) h.fn();
  }

  function onMouseMove(e) {
    const {x, y} = canvasPos(e);
    const h = hitAt(x, y);
    const id = h ? h.id : null;
    if (id !== S.hoverId) {
      S.hoverId = id;
      S.canvas.style.cursor = h ? 'pointer' : 'default';
      render();
    }
  }

  function onWheel(e) {
    const pd = S.pd; if (!pd) return;
    const dy = e.deltaY;
    if (pd.page === 'game') {
      S.movesScroll = Math.max(0, Math.min(Math.max(0, S.movesContentH - 100), S.movesScroll + dy));
      render();
    } else if (pd.page === 'index') {
      S.gamesScroll = Math.max(0, Math.min(Math.max(0, S.gamesContentH - 100), S.gamesScroll + dy));
      render();
    }
  }

  function onTouchStart(e) {
    if (e.touches.length !== 1) return;
    S._tY0 = e.touches[0].clientY;
    S._tS0 = (S.pd && S.pd.page === 'game') ? S.movesScroll : S.gamesScroll;
  }

  function onTouchMove(e) {
    if (e.touches.length !== 1) return;
    const dy = S._tY0 - e.touches[0].clientY;
    const pd = S.pd; if (!pd) return;
    if (pd.page === 'game') {
      S.movesScroll = Math.max(0, Math.min(Math.max(0, S.movesContentH - 100), S._tS0 + dy));
      render();
    } else if (pd.page === 'index') {
      S.gamesScroll = Math.max(0, Math.min(Math.max(0, S.gamesContentH - 100), S._tS0 + dy));
      render();
    }
  }

  function onTouchEnd(e) {
    if (e.changedTouches.length !== 1) return;
    const dy = Math.abs(e.changedTouches[0].clientY - S._tY0);
    if (dy < 12) {
      const r = S.canvas.getBoundingClientRect();
      const cx = e.changedTouches[0].clientX - r.left;
      const cy = e.changedTouches[0].clientY - r.top;
      const h = hitAt(cx, cy);
      if (h) h.fn();
    }
  }

  function onKeydown(e) {
    const pd = S.pd;
    if (!pd || pd.page !== 'game') return;
    const tag = e.target && e.target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    if (e.key === 'ArrowRight' || e.key === 'j') { e.preventDefault(); gotoPlay(pd, S.ply + 1); }
    else if (e.key === 'ArrowLeft' || e.key === 'k') { e.preventDefault(); gotoPlay(pd, S.ply - 1); }
    else if (e.key === 'Home')  { e.preventDefault(); gotoPlay(pd, 0); }
    else if (e.key === 'End')   { e.preventDefault(); gotoPlay(pd, pd.max_ply); }
  }

  // ── Resize ────────────────────────────────────────────────────────────────
  function resize() {
    S.dpr = window.devicePixelRatio || 1;
    S.W = window.innerWidth;
    S.H = window.innerHeight;
    S.canvas.width  = Math.round(S.W * S.dpr);
    S.canvas.height = Math.round(S.H * S.dpr);
    S.canvas.style.width  = S.W + 'px';
    S.canvas.style.height = S.H + 'px';
    render();
  }

  // ── Boot ──────────────────────────────────────────────────────────────────
  function init() {
    const dataEl = document.getElementById('page-data');
    if (!dataEl) return;
    S.pd = JSON.parse(dataEl.textContent);

    if (S.pd.page === 'game') { S.ply = S.pd.ply; S.fen = S.pd.fen; }
    if (S.pd.page === 'import') { S.pgnText = S.pd.text || ''; }

    S.canvas = document.getElementById('app');
    S.ctx = S.canvas.getContext('2d');

    S.canvas.addEventListener('click', onClickCanvas);
    S.canvas.addEventListener('mousemove', onMouseMove);
    S.canvas.addEventListener('wheel', onWheel, {passive: true});
    S.canvas.addEventListener('touchstart', onTouchStart, {passive: true});
    S.canvas.addEventListener('touchmove', onTouchMove, {passive: true});
    S.canvas.addEventListener('touchend', onTouchEnd);
    document.addEventListener('keydown', onKeydown);

    // Overlay input
    const ovrInp = document.getElementById('ovr-input');
    if (ovrInp) {
      ovrInp.addEventListener('blur', deactivateOvr);
      ovrInp.addEventListener('keydown', e => { if (e.key === 'Enter') ovrInp.blur(); });
    }

    // PGN textarea
    const ovrPgn = document.getElementById('ovr-pgn');
    if (ovrPgn) {
      ovrPgn.addEventListener('blur', deactivatePgn);
    }

    // File input (import page only)
    const impFile = document.getElementById('import-file');
    if (impFile) {
      impFile.addEventListener('change', () => {
        if (impFile.files.length) document.getElementById('import-form').submit();
      });
    }

    new ResizeObserver(resize).observe(document.body);
    window.addEventListener('orientationchange', () => setTimeout(resize, 100));
    resize();
    loadPieces().then(render);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
