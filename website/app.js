// app.js
// IntersectionObserver for reveal effects
const io = new IntersectionObserver(
  (entries) => {
    for (const e of entries) {
      if (e.isIntersecting) {
        e.target.classList.add('in-view');
        io.unobserve(e.target);
      }
    }
  },
  { threshold: 0.12 }
);

document.querySelectorAll('.reveal').forEach(el => io.observe(el));

// Parallax on scroll using data attributes
const parallaxEls = Array.from(document.querySelectorAll('[data-parallax]'));
let lastY = window.scrollY;
function parallax() {
  const y = window.scrollY;
  const dy = y - lastY;
  lastY = y;
  parallaxEls.forEach(el => {
    const speed = parseFloat(el.getAttribute('data-parallax-speed') || '0.15');
    const rect = el.getBoundingClientRect();
    // Only transform while near viewport
    if (rect.top < window.innerHeight && rect.bottom > 0) {
      el.style.transform = `translate3d(0, ${y * speed * -0.2}px, 0)`;
    }
  });
  requestAnimationFrame(parallax);
}
requestAnimationFrame(parallax);

// Card hover spotlight position
document.querySelectorAll('.card').forEach(card => {
  card.addEventListener('mousemove', (e) => {
    const r = card.getBoundingClientRect();
    const mx = e.clientX - r.left;
    const my = e.clientY - r.top;
    card.style.setProperty('--mx', `${mx}px`);
    card.style.setProperty('--my', `${my}px`);
  });
});

// Mobile nav toggle
const toggle = document.querySelector('.nav-toggle');
const links = document.querySelector('.nav-links');
if (toggle && links) {
  toggle.addEventListener('click', () => {
    const visible = links.style.display === 'flex';
    links.style.display = visible ? 'none' : 'flex';
  });
}

// Canvas colorful accent without breaking B/W theme
const canvas = document.getElementById('fx-canvas');
const ctx = canvas.getContext('2d', { alpha: true });
let dpr = Math.max(1, Math.min(2, window.devicePixelRatio || 1));
let W = 0, H = 0;

function resize() {
  W = canvas.clientWidth = window.innerWidth;
  H = canvas.clientHeight = window.innerHeight;
  canvas.width = Math.floor(W * dpr);
  canvas.height = Math.floor(H * dpr);
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
}
resize();
window.addEventListener('resize', resize);

const N = 18;
const orbs = Array.from({ length: N }).map((_, i) => ({
  x: Math.random() * W,
  y: Math.random() * H,
  r: 40 + Math.random() * 120,
  a: Math.random() * Math.PI * 2,
  va: 0.002 + Math.random() * 0.004,
  hue: (i * (360 / N) + Math.random() * 30) % 360,
  speed: 0.4 + Math.random() * 0.6
}));

function tick() {
  ctx.clearRect(0, 0, W, H);
  for (const o of orbs) {
    o.a += o.va;
    o.x += Math.cos(o.a) * o.speed;
    o.y += Math.sin(o.a * 0.8) * o.speed;

    // wrap
    if (o.x < -200) o.x = W + 200;
    if (o.x > W + 200) o.x = -200;
    if (o.y < -200) o.y = H + 200;
    if (o.y > H + 200) o.y = -200;

    // subtle colorful glow
    const grd = ctx.createRadialGradient(o.x, o.y, 0, o.x, o.y, o.r);
    grd.addColorStop(0, `hsla(${o.hue}, 85%, 60%, 0.12)`);
    grd.addColorStop(1, `hsla(${o.hue}, 85%, 60%, 0.0)`);
    ctx.fillStyle = grd;
    ctx.beginPath();
    ctx.arc(o.x, o.y, o.r, 0, Math.PI * 2);
    ctx.fill();
  }
  requestAnimationFrame(tick);
}
requestAnimationFrame(tick);

// --- Lightbox logic for screenshots ---
function initLightbox() {
  const lightbox = document.getElementById('lightbox');
  const lightboxImg = document.querySelector('.lightbox-img');
  const lightboxClose = document.querySelector('.lightbox-close');

  function openLightbox(src, alt = '') {
    if (!lightbox || !lightboxImg) return;
    lightboxImg.src = src;
    lightboxImg.alt = alt;
    lightbox.classList.add('open');
    lightbox.setAttribute('aria-hidden', 'false');
    document.body.style.overflow = 'hidden';
    lightboxClose && lightboxClose.focus();
  }

  function closeLightbox() {
    if (!lightbox || !lightboxImg) return;
    lightbox.classList.remove('open');
    lightbox.setAttribute('aria-hidden', 'true');
    document.body.style.overflow = '';
    lightboxImg.src = '';
    lightboxImg.alt = '';
  }

  // Robust binding: attach directly to each screenshot image
  const imgs = document.querySelectorAll('#screenshots .media-card img');
  imgs.forEach(img => {
    // Accessibility affordances
    img.setAttribute('tabindex', '0');
    img.setAttribute('role', 'button');
    img.setAttribute('aria-label', (img.alt || 'Open screenshot') + ' – open preview');

    const open = () => openLightbox(img.currentSrc || img.src, img.alt || '');
    img.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      open();
    });
    img.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        open();
      }
    });
  });

  // Close events
  lightboxClose && lightboxClose.addEventListener('click', closeLightbox);
  lightbox && lightbox.addEventListener('click', (e) => {
    if (e.target === lightbox) closeLightbox();
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeLightbox();
  });
}

// Ensure DOM is ready before binding
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initLightbox);
} else {
  initLightbox();
}
