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
    let Rules = [];
    let Edit = null;

    function refresh() {
    fetch("/api/rules").then(t=>t.json()).then(data=>{
        Rules = data
    });
  }

  function set(ele) {
  }

  function del(id) {
    var formData = new FormData();
    formData.append("ID", id)

    fetch("/api/ruleDel",{
      method: 'POST',
      body: formData,
    }).then(t=>t.text()).then(d=>{
      this.refresh()
    });
  }

  function doSave() {
    var formData = new FormData();
    formData.append("Items", Edit.Items)
    formData.append("Note", Edit.Note)
    formData.append("Proxy", Edit.Proxy ? "1":"")
    formData.append("Enable", Edit.Enable ? "1":"")

    let url

    if (edit == null) {
      url = "/api/ruleAdd";
    } else {
      url = "/api/ruleEdit";
      formData.append("ID", this.state.edit.ID)
    }

    fetch(url,{
      method: 'POST',
      body: formData,
    }).then(t=>t.text()).then(d=>{
      this.refresh()
    });
  }
</script>

<svelte:head>
    <title>index</title>
</svelte:head>

<div>
      <div className={styles.rulesNav}>
      {Rules.map(r=>(r.ID===this.state.edit.ID) ?
        <span><b>{r.Note}</b></span>
       :
        <span onclick={()=>set(r)}>{r.Note}</span>
      )}
      &nbsp;
    <span onclick={()=>set()}>Add</span>
      </div>
      <div>
      <div className={styles.tr}>
      <span className={styles.td}>Note</span>
      <input value={edit.Note} onchange={c1} />
      </div>
      <div className={styles.tr}>
      <span className={styles.td}>Action:</span>
        <label>
        <input type="radio" name="proxy" checked={edit.Proxy} onchange={c2} /> Proxy
        </label>
        <label>
        <input type="radio" name="proxy" checked={!edit.Proxy} onchange={c2} /> Direct
        </label>
      </div>
      <div className={styles.tr}>
      <span className={styles.td}>Enable</span>
      <input type="checkbox" checked={edit.Enable} onchange={c3} />
      </div>
      <div>
      <textarea className={styles.td} value={edit.Items} onchange={c4} />
      </div>
      <div>
      <input type="button" value="Save" onclick={()=>this.doSave()} /> &nbsp;
    {edit.ID !== "" &&
        <input type="button" value="Del" onclick={()=>this.del(edit.ID)} />
    }
      </div>
      </div>
      </div>

<style>
    main {
        color: red;
    }
</style>
