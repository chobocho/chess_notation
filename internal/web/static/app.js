(function () {
  'use strict';

  const PIECE_CODES = ['wK','wQ','wR','wB','wN','wP','bK','bQ','bR','bB','bN','bP'];
  const pieceImages = {};
  let piecesReady = null;

  function loadPieces() {
    if (piecesReady) return piecesReady;
    piecesReady = Promise.all(PIECE_CODES.map(code => new Promise((resolve) => {
      const img = new Image();
      img.decoding = 'async';
      img.onload = () => { pieceImages[code] = img; resolve(); };
      img.onerror = () => resolve();
      img.src = '/static/pieces/' + code + '.svg';
    })));
    return piecesReady;
  }

  function parseFEN(fen) {
    const board = Array.from({ length: 8 }, () => new Array(8).fill(null));
    if (!fen) return board;
    const rows = fen.split(' ')[0].split('/');
    if (rows.length !== 8) return board;
    for (let i = 0; i < 8; i++) {
      let file = 0;
      for (const c of rows[i]) {
        if (c >= '1' && c <= '8') { file += c.charCodeAt(0) - 48; continue; }
        const color = (c >= 'A' && c <= 'Z') ? 'w' : 'b';
        const letter = c.toUpperCase();
        const rank = 7 - i;
        if (file < 8) board[rank][file] = color + letter;
        file++;
      }
    }
    return board;
  }

  function drawBoard(canvas) {
    const fen = canvas.dataset.fen || '';
    const rect = canvas.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    const sizeCSS = Math.max(128, Math.floor(rect.width));
    canvas.width = Math.floor(sizeCSS * dpr);
    canvas.height = Math.floor(sizeCSS * dpr);
    canvas.style.height = sizeCSS + 'px';

    const ctx = canvas.getContext('2d');
    ctx.save();
    ctx.scale(dpr, dpr);
    const sq = sizeCSS / 8;
    const light = '#f0d9b5', dark = '#b58863';
    const board = parseFEN(fen);

    for (let r = 7; r >= 0; r--) {
      for (let f = 0; f < 8; f++) {
        const x = f * sq;
        const y = (7 - r) * sq;
        ctx.fillStyle = ((f + r) % 2 === 0) ? dark : light;
        ctx.fillRect(x, y, sq, sq);
      }
    }

    // Pieces.
    for (let r = 7; r >= 0; r--) {
      for (let f = 0; f < 8; f++) {
        const piece = board[r][f];
        if (!piece) continue;
        const img = pieceImages[piece];
        if (!img || !img.complete) continue;
        const x = f * sq;
        const y = (7 - r) * sq;
        const pad = sq * 0.06;
        ctx.drawImage(img, x + pad, y + pad, sq - 2 * pad, sq - 2 * pad);
      }
    }

    // Coordinate labels (file a-h on bottom, rank 1-8 on right).
    ctx.font = Math.max(9, Math.round(sq * 0.18)) + 'px system-ui, sans-serif';
    for (let f = 0; f < 8; f++) {
      const x = f * sq;
      ctx.fillStyle = ((f + 0) % 2 === 0) ? '#f0d9b5' : '#b58863';
      ctx.textBaseline = 'bottom';
      ctx.textAlign = 'left';
      ctx.fillText(String.fromCharCode(97 + f), x + sq * 0.06, sizeCSS - sq * 0.06);
    }
    for (let r = 0; r < 8; r++) {
      const y = (7 - r) * sq;
      ctx.fillStyle = ((7 + r) % 2 === 0) ? '#f0d9b5' : '#b58863';
      ctx.textBaseline = 'top';
      ctx.textAlign = 'right';
      ctx.fillText(String(r + 1), sizeCSS - sq * 0.06, y + sq * 0.06);
    }
    ctx.restore();
  }

  function scheduleRedraw(canvas) {
    if (canvas._redrawScheduled) return;
    canvas._redrawScheduled = true;
    requestAnimationFrame(() => {
      canvas._redrawScheduled = false;
      drawBoard(canvas);
    });
  }

  function initCanvas(canvas) {
    loadPieces().then(() => drawBoard(canvas));
    const ro = new ResizeObserver(() => scheduleRedraw(canvas));
    ro.observe(canvas);
    window.addEventListener('orientationchange', () => scheduleRedraw(canvas));
  }

  function initViewer() {
    const canvas = document.querySelector('canvas.board-canvas');
    if (!canvas) return;
    initCanvas(canvas);

    const section = document.querySelector('section.game');
    if (!section) return;
    const gameID = section.dataset.gameId;
    let ply = parseInt(section.dataset.ply, 10) || 0;
    const max = parseInt(section.dataset.max, 10) || 0;
    const indicator = document.getElementById('ply-indicator');

    function goto(n) {
      n = Math.max(0, Math.min(max, n));
      if (n === ply) return;
      fetch('/game/' + gameID + '/fragment/' + n, { headers: { 'Accept': 'text/plain' } })
        .then(r => r.ok ? r.text() : null)
        .then(fen => {
          if (fen == null) return;
          canvas.dataset.fen = fen.trim();
          drawBoard(canvas);
          ply = n;
          if (indicator) indicator.textContent = n + ' / ' + max;
          history.replaceState(null, '', '/game/' + gameID + '/ply/' + n);
          document.querySelectorAll('ol.moves li').forEach((li, idx) => {
            li.classList.toggle('active', idx + 1 === n);
          });
        });
    }

    document.addEventListener('keydown', function (e) {
      const tag = e.target && e.target.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;
      if (e.key === 'ArrowRight' || e.key === 'j') { e.preventDefault(); goto(ply + 1); }
      else if (e.key === 'ArrowLeft' || e.key === 'k') { e.preventDefault(); goto(ply - 1); }
      else if (e.key === 'Home') { e.preventDefault(); goto(0); }
      else if (e.key === 'End') { e.preventDefault(); goto(max); }
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initViewer);
  } else {
    initViewer();
  }
})();
