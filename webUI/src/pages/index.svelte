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
    export let a = 0;
</script>

<svelte:head>
    <title>index</title>
</svelte:head>

<main>
    index
</main>
{a}

<style>
    main {
        color: red;
    }
</style>
