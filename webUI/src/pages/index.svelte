<script>
       import Layout from "./lib/layout.svelte";
       import State from "./lib/state.svelte";
       import Logs from "./lib/logs.svelte";

       let type=1
</script>

<Layout>
<div>
    <div>
        <a on:click="{()=>{type=1}}">now</a>
        <a on:click="{()=>{type=0}}">log</a>
    </div>
    <State />
    <Logs type={type} />
</div>
</Layout>

<style>
</style>
