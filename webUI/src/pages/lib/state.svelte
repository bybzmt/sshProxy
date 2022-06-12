
<script>
  import { onMount, onDestroy } from 'svelte';

  let state={};

      function load() {
            fetch(API_BASE + "/api/state").then(t=>t.json()).then(data=>{
                state = data
          });
    }

    let timer;

    onMount(() => {
        timer = setInterval(load, 500);
    });

   onDestroy(()=>{
        clearInterval(timer);
    });
</script>

<div>
          <span>ConnNum: { state.ConnNum }</span> &nbsp;
          <span>Incoming: { state.Incoming }</span> &nbsp;
          <span>Outgoing: { state.Outgoing }</span>
        </div>

<style>
</style>