document.addEventListener('DOMContentLoaded', () => {
  const here = location.pathname.toLowerCase();
  const links = document.querySelectorAll('.sidebar nav a[href]');
  let best = null;
  links.forEach(a => {
    const href = a.getAttribute('href');
    if (!href || href === '#' || href.startsWith('javascript')) return;
    try {
      const url = new URL(href, location.origin);
      const path = url.pathname.toLowerCase();
      if (path === here) {
        a.classList.add('active');
      } else if (!best && here.startsWith(path) && path !== '/') {
        best = a; // fallback parcial
      }
    } catch (_) {}
  });
  if (!document.querySelector('.sidebar nav a.active') && best) {
    best.classList.add('active');
  }
});

