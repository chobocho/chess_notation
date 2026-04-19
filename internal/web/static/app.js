(function () {
  const section = document.querySelector('section.game');
  if (!section) return;
  const gameID = section.dataset.gameId;
  let ply = parseInt(section.dataset.ply, 10) || 0;
  const max = parseInt(section.dataset.max, 10) || 0;
  const boardEl = document.getElementById('board');
  const indicator = document.getElementById('ply-indicator');

  function render(n) {
    n = Math.max(0, Math.min(max, n));
    if (n === ply) return;
    fetch(`/game/${gameID}/fragment/${n}`)
      .then(r => r.ok ? r.text() : null)
      .then(html => {
        if (html == null) return;
        boardEl.innerHTML = html;
        ply = n;
        indicator.textContent = `${n} / ${max}`;
        history.replaceState(null, '', `/game/${gameID}/ply/${n}`);
        document.querySelectorAll('ol.moves li').forEach((li, idx) => {
          li.classList.toggle('active', idx + 1 === n);
        });
      });
  }

  document.addEventListener('keydown', function (e) {
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
    if (e.key === 'ArrowRight' || e.key === 'j') { e.preventDefault(); render(ply + 1); }
    else if (e.key === 'ArrowLeft' || e.key === 'k') { e.preventDefault(); render(ply - 1); }
    else if (e.key === 'Home') { e.preventDefault(); render(0); }
    else if (e.key === 'End') { e.preventDefault(); render(max); }
  });
})();
