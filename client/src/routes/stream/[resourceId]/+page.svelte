<script>
  import { onMount, onDestroy } from "svelte";
  import Hls from "hls.js";

  export let data;
  let source = `http://localhost:8080/stream/${data.resourceId}/hls/stream.m3u8`;

  let videoElement;
  let hls;

  onMount(() => {
    if (Hls.isSupported()) {
      hls = new Hls();
      hls.loadSource(source);
      hls.attachMedia(videoElement);
    } else if (videoElement.canPlayType("application/vnd.apple.mpegurl")) {
      videoElement.src = source;
    }
  });

  onDestroy(() => {
    if (hls) {
      hls.destroy();
    }
  });
</script>

<svelte:head>
  <title>"Watch neralai"</title>
</svelte:head>

<!-- svelte-ignore a11y-media-has-caption -->
<video bind:this={videoElement} autoplay controls />

<style>
  :global(video) {
    width: 100vw;
    height: 100vh;
    object-fit: cover;
  }
</style>
