<script>
    import Layout from "./lib/layout.svelte";
    import Ciphers from "./lib/ciphers.svelte";
    import { onMount } from "svelte";

    let Servers = [];
    let editDefault = {
        Addr: "",
        Passwd: "",
        Cipher: "",
        Note: "",
        Enable: true,
    };
    let edit = editDefault;

    function load() {
        fetch(API_BASE + "/api/serverConfigs")
            .then((t) => t.json())
            .then((data) => {
                Servers = data;
            });
    }

    onMount(() => {
        load();
    });

    let save = (data) => {
        var formData = new FormData();
        formData.append("ID", data.ID);
        formData.append("Addr", data.Addr);
        formData.append("Passwd", data.Passwd);
        formData.append("Cipher", data.Cipher);
        formData.append("Note", data.Note);
        formData.append("Enable", data.Enable);

        let url;
        if (data.ID) {
            url = "/api/serverConfigEdit";
        } else {
            url = "/api/serverConfigAdd";
        }

        fetch(API_BASE + url, {
            method: "POST",
            body: formData,
        })
            .then((t) => t.text())
            .then((d) => {
                load();

                console.log(editDefault);
                edit = editDefault;
            });
    };

    let del = (data) => {
        var formData = new FormData();
        formData.append("ID", data.ID);

        fetch(API_BASE + "/api/serverConfigDel", {
            method: "POST",
            body: formData,
        })
            .then((t) => t.text())
            .then((d) => {
                load();
            });
    };
</script>

<Layout>
    <table>
        <tr>
            <th>ID</th>
            <th>Addr</th>
            <th>Cipher</th>
            <th>User</th>
            <th>Password</th>
            <th>Note</th>
            <th>Enable</th>
            <th class="w-20" />
        </tr>

        {#each Servers as server}
            <tr>
                <td>{server.ID}</td>
                <td><input class="border w-full" bind:value={server.Addr} /></td
                >
                <td>
                    <Ciphers bind:value={server.Cipher} />
                </td>
                <td><input class="border w-full" bind:value={edit.User} /></td>
                <td
                    ><input
                        class="border w-full"
                        bind:value={server.Passwd}
                    /></td
                >

                <td><input class="border w-full" bind:value={server.Note} /></td
                >
                <td
                    ><input
                        class="border w-full"
                        type="checkbox"
                        bind:value={server.Enable}
                    /></td
                >
                <td>
                    <button
                        type="button"
                        on:click={() => {
                            save(server);
                        }}>edit</button
                    >
                    &nbsp;
                    <button
                        type="button"
                        on:click={() => {
                            del(server);
                        }}>del</button
                    >
                </td>
            </tr>
        {/each}
        <tr>
            <td>--</td>
            <td><input class="border w-full" bind:value={edit.Addr} /></td>
            <td>
                <Ciphers bind:value={edit.Cipher} />
            </td>
            <td><input class="border w-full" bind:value={edit.User} /></td>
            <td><input class="border w-full" bind:value={edit.Passwd} /></td>
            <td><input class="border w-full" bind:value={edit.Note} /></td>
            <td
                ><input
                    class="border w-full"
                    type="checkbox"
                    bind:value={edit.Enable}
                /></td
            >
            <td
                ><button
                    type="button"
                    on:click={() => {
                        save(edit);
                    }}>add</button
                ></td
            >
        </tr>
    </table>
</Layout>

<style>
</style>
