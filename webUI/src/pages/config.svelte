<script context="module">
    import Layout from "./layout.sevlte";

    export const load = async () => {
        let data = await fetch("/api/clientConfig").then((t) => t.json());

        return {
            props: {
                data: data,
            },
        };
    };
</script>

<script>
    export let data;

    function doSave() {
        var formData = new FormData();
        formData.append("Addr", data.Addr);
        formData.append("Timeout", data.Timeout);
        formData.append("IdleTimeout", data.IdleTimeout);
        formData.append("Proxy", data.Proxy ? "1" : "");
        formData.append("LDNS", data.LDNS);
        formData.append("RDNS", data.RDNS);

        fetch("/api/clientConfigSave", {
            method: "POST",
            body: formData,
        })
            .then((t) => t.json())
            .then((d) => {
                this.setState(d);
                fetch("/api/restart");
            });
    }
</script>

<Layout>
    <form>
        <div class="tr">
            <span>Addr:</span> &nbsp; <input bind:value={data.Addr} />
        </div>
        <div class="tr">
            <span>Timeout:</span> &nbsp; <input bind:value={data.Timeout} />
        </div>
        <div class="tr">
            <span>IdleTimeout:</span> &nbsp; <input bind:value={data.IdleTimeout} />
        </div>
        <div class="tr">
            <span>Default Action:</span> &nbsp;

            <label>
                <input type="radio" name="proxy" bind:group={data.Proxy} value="1" />
                Proxy
            </label>

            <label>
                <input type="radio" name="proxy" bind:group={data.Proxy} value="0" />
                Direct
            </label>
        </div>
        <div class="tr">
            <span>DNS(on Direct):</span> &nbsp; <textarea bind:value={data.LDNS} />
        </div>
        <div class="tr">
            <span>RDNS(on Proxy):</span> &nbsp; <textarea bind:value={data.RDNS} />
        </div>
        <div class="tr">
            <input type="button" value="save & restart" on:click={doSave} />
        </div>
    </form>
</Layout>

<style>
    .tr {
        display: table-row;
    }
    .td {
        display: table-cell;
    }
</style>
