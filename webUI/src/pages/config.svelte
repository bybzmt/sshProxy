<script>
    import Layout from "./lib/layout.svelte";
    import { onMount } from "svelte";

    let data = {
        Addr: "",
        Timeout: 0,
        IdleTimeout: 0,
        LDNS: "",
        LDNSEnable: false,
        RDNS: "",
        RDNSEnable: false,
    };

    function load() {
        fetch(API_BASE + "/api/clientConfig")
            .then((t) => t.json())
            .then((d) => {
                data = d;
            });
    }

    function doSave() {
        var formData = new FormData();
        formData.append("Addr", data.Addr);
        formData.append("Timeout", data.Timeout);
        formData.append("IdleTimeout", data.IdleTimeout);
        formData.append("LDNS", data.LDNS);
        formData.append("LDNSEnable", data.LDNSEnable ? "1" : "");
        formData.append("RDNS", data.RDNS);
        formData.append("RDNSEnable", data.RDNSEnable ? "1" : "");

        fetch(API_BASE + "/api/clientConfigSave", {
            method: "POST",
            body: formData,
        })
            .then((t) => t.json())
            .then((d) => {
                load();
                fetch(API_BASE + "/api/restart");
            });
    }

    onMount(() => {
        load();
    });
</script>

<Layout>
    <form>
        <table>
            <tr>
                <td><span>Addr:</span></td>
                <td><input class="border" bind:value={data.Addr} /></td>
            </tr>
            <tr>
                <td><span>Timeout:</span> </td>
                <td><input class="border" bind:value={data.Timeout} /></td>
            </tr>
            <tr>
                <td><span>IdleTimeout:</span></td>
                <td><input class="border" bind:value={data.IdleTimeout} /></td>
            </tr>
            <tr>
                <td class="align-top">
                    <span>DNS:<br />(on Direct)</span>
                </td>
                <td>
                    <textarea class="border" bind:value={data.LDNS} />
                    <br />
                    <label>
                        <input type="radio" name="LDNSEable" bind:group={data.LDNSEnable} value={true} />
                        Enable
                    </label>

                    <label>
                        <input type="radio" name="LDNSEable" bind:group={data.LDNSEnable} value={false} />
                        Disable
                    </label>
                </td>
            </tr>
            <tr>
                <td class="align-top">
                    <span>RDNS:<br />(on Proxy)</span>
                </td>
                <td
                    ><textarea class="border" bind:value={data.RDNS} />
                    <br />
                    <label>
                        <input type="radio" name="RDNSEable" bind:group={data.RDNSEnable} value={true} />
                        Enable
                    </label>

                    <label>
                        <input type="radio" name="RDNSEable" bind:group={data.RDNSEnable} value={false} />
                        Disable
                    </label>
                </td>
            </tr>
            <tr>
                <td colspan="2" class="text-right">
                    <button class="border" type="button" on:click={doSave}>save & restart</button>
                </td>
            </tr>
        </table>
    </form>
</Layout>

<style>
</style>
