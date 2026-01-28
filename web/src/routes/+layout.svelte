<script lang="ts">
  import '../app.css';
  import NavBar from '../components/NavBar.svelte';
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/service-worker.js')
      .then((registration) => {
        console.log('Service Worker registered with scope:', registration.scope);
      })
      .catch((error) => {
        console.error('Service Worker registration failed:', error);
      });
  }
</script>

<NavBar/>

<main class="container mx-auto px-4 py-6 min-h-screen">
  <slot />
</main>

<footer class="footer footer-center bg-base-300 text-base-content p-4 mt-auto border-t border-base-content/10">
  <aside>
    <p class="text-sm opacity-70">© 2024 DKP Auction House - All rights reserved</p>
  </aside>
</footer>
