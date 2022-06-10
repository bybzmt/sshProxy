<script context="module">
    export const load = async () => {
        console.log("load1");

        return {
            props: {
                a: Math.random(),
            },
        };
    };
</script>

<script>
    import { onMount, onDestroy } from 'svelte';

    export let logs = [];
    export let type = 0;
    export let length = 100;

    function load() {
        let url = "/api/log?type=" + type + "&length=" + length;

        fetch(url)
            .then((t) => t.json())
            .then((d) => {
                logs = d;
            });
    }

    let timer;
    onMount(() => {
        timer = setInterval(load, 500);
    }

    onDestroy(()=>{
        clearInterval(timer);
    })
</script>

<table>
    {#each logs as log}
        <tr>
            <td>{log.ID}</td>
            <td>{new Date(log.Now).toLocaleString()}</td>
            <td>{log.Proxy ? "Proxy" : "Direct"}</td>
            <td>{log.From}</td>
            <td>{log.To}</td>
            <td>{log.Msg}</td>
        </tr>
    {/each}
</table>

<style>
    table {
        border-collapse: collapse;
    }
    table td,
    table th {
        border: 1px solid #777;
        padding: 3px 5px 2px;
    }
</style>
