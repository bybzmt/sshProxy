<script>
    import "./tailwind.css";
</script>

<div>
    <nav>
        <a href="#/">Index</a>
        <a href="#/state">State</a>
        <a href="#/config">Client</a>
        <a href="#/server">Server</a>
        <a href="#/rules">Rules</a>
    </nav>

    <slot />
</div>

<style>
    nav {
        margin: 1em 0;
    }
    nav a {
        margin: 0 1em 0 0;
    }
    div {
        margin: 1em;
    }
</style>
