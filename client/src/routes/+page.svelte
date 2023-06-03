<script>
  import { onMount } from "svelte";
  import { WHIPClient } from "@eyevinn/whip-web-client";
  import { copy } from "svelte-copy";

  let goLive = false;
  let videoIngest;
  let mediaStream;
  let client;
  let resourceId;
  let streamLink;

  onMount(async () => {
    client = new WHIPClient({
      endpoint: "http://localhost:8080/stream",
      opts: {
        debug: true,
        iceServers: [{ urls: "stun:stun.l.google.com:19320" }],
      },
    });
    client.setIceServersFromEndpoint();

    videoIngest = document.querySelector("video#ingest");
  });

  async function toggleLive() {
    if (!goLive) {
      try {
        mediaStream = await navigator.mediaDevices.getUserMedia({
          // highest possible resolution from camera
          video: {
            width: { ideal: 4096 },
            height: { ideal: 2160 },
          },
          audio: true,
        });
        videoIngest.srcObject = mediaStream;
        await client.ingest(mediaStream);
        let resourceUrl = await client.getResourceUrl();
        resourceId = resourceUrl.substring(resourceUrl.lastIndexOf("/") + 1);
        streamLink = `${window.location.origin}/stream/${resourceId}`;
        goLive = true;
      } catch (err) {
        console.error(err);
      }
    } else {
      await client.destroy();
      mediaStream.getTracks().forEach((track) => track.stop());
      goLive = false;
    }
  }
</script>

<svelte:head>
  <title>"Start your neralai"</title>
</svelte:head>

<div id="center">
  <img src="/neralai.png" alt="Neralai" id="logo" />
  <div id="videoContainer">
    <video id="ingest" autoplay muted />
  </div>
  <button id="goLive" on:click={toggleLive}>
    {goLive ? "Stop Live" : "Go Live"}
  </button>
  {#if goLive}
    <div id="linkContainer">
      {#await new Promise((res) => setTimeout(res, 4000))}
        <p>Loading stream...</p>
      {:then val}
        <a href="/stream/{resourceId}" target="_blank">Watch here</a>
        <div class="copyLink">
          <input type="text" bind:value={streamLink} readonly />
          <button use:copy={streamLink}>Copy link</button>
        </div>
      {/await}
    </div>
  {/if}
</div>

<style>
  #center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    margin: 0;
    padding: 0;
    font-family: Arial, sans-serif;
  }

  #videoContainer {
    position: relative;
    width: 320px;
    height: 240px;
    margin-bottom: 20px;
  }

  #ingest {
    width: 100%;
    height: 100%;
  }

  #goLive {
    padding: 10px 20px;
    background-color: #4caf50;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 16px;
    cursor: pointer;
  }

  .copyLink {
    display: flex;
    align-items: center;
    margin-top: 10px;
  }

  .copyLink input[type="text"] {
    flex: 1;
    padding: 5px;
    font-size: 14px;
    border: none;
    border-radius: 4px;
    background-color: #f2f2f2;
  }

  .copyLink button {
    padding: 5px 10px;
    margin-left: 10px;
    background-color: #4caf50;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    cursor: pointer;
  }
</style>
