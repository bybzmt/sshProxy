<script>
  import Layout from "./lib/layout.svelte";
  import { onMount } from "svelte";

  let Rules = [];
  let Edit = {
    Note: "",
    Enable: false,
    Items: "",
  };

  function refresh() {
    fetch(API_BASE + "/api/rules")
      .then((t) => t.json())
      .then((data) => {
        Rules = data;
      });
  }

  function del(data) {
    var formData = new FormData();
    formData.append("ID", data.ID);

    fetch(API_BASE + "/api/ruleDel", {
      method: "POST",
      body: formData,
    })
      .then((t) => t.text())
      .then((d) => {
        refresh();
      });
  }

  function doSave(data) {
    var formData = new FormData();
    formData.append("Items", data.Items);
    formData.append("Note", data.Note);
    formData.append("Servers", data.Servers);
    formData.append("Enable", data.Enable ? "1" : "");
    formData.append("ID", data.ID);

    let url;

    if (data.ID) {
      url = "/api/ruleEdit";
    } else {
      url = "/api/ruleAdd";
    }

    fetch(API_BASE + url, {
      method: "POST",
      body: formData,
    })
      .then((t) => t.text())
      .then((d) => {
        refresh();
      });
  }

  onMount(() => {
    refresh();
  });
</script>

<Layout>
  <table>
    <tr>
      <td>ID</td>
      <td>Note</td>
      <td>Roules</td>
      <td>Servers</td>
      <td>Enable</td>
      <td />
    </tr>

    {#each Rules as rule}
      <tr>
        <td>{rule.ID}</td>
        <td><input class="border w-full" bind:value={rule.Note} /></td>
        <td>
          <textarea class="border w-full" bind:value={rule.Items} />
        </td>
        <td><input class="border w-full" bind:value={rule.Servers} /></td>
        <td>
          <input type="checkbox" bind:checked={rule.Enable} />
        </td>
        <td>
          <button class="border" type="button" on:click={() => doSave(rule)}>Save</button>
          <button class="border" type="button" on:click={() => del(rule)}>Del</button>
        </td>
      </tr>
    {/each}

    <tr>
      <td>--</td>
      <td><input class="border w-full" bind:value={Edit.Note} /></td>
      <td>
        <textarea class="border w-full" bind:value={Edit.Items} />
      </td>
      <td><input class="border w-full" bind:value={Edit.Servers} /></td>
      <td>
        <input type="checkbox" bind:checked={Edit.Enable} />
      </td>
      <td>
        <button class="border" type="button" on:click={() => doSave(Edit)}>Add</button>
      </td>
    </tr>
  </table>
</Layout>

<style>
  td {
    vertical-align: top;
  }
</style>
