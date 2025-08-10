<script>
  import { onMount } from "svelte";

  // Basic widget builder using Svelte 5 runes. Focused, clean UI.
  let templates = $state([]);
  let selectedTemplate = $state(null);
  let widget = $state({
    name: "",
    api_url: "",
    api_headers: {},
    description: "",
    timeout: 30,
    enabled: true
  });
  let mapping = $state({});
  let loading = $state(false);
  let error = $state("");
  let success = $state("");

  async function apiGet(endpoint) {
    const r = await fetch("/api" + endpoint);
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }
  async function apiPost(endpoint, body) {
    const r = await fetch("/api" + endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    });
    if (!r.ok) throw new Error("HTTP " + r.status);
    return r.json();
  }

  async function loadTemplates() {
    try {
      const data = await apiGet("/widgets/templates");
      templates = data.templates || [];
    } catch (e) {
      error = "Failed to load templates: " + e.message;
    }
  }

  async function save() {
    if (!selectedTemplate || !widget.name) {
      error = "Please provide a widget name and choose a template.";
      return;
    }
    try {
      loading = true;
      error = "";
      success = "";
      const payload = {
        ...widget,
        template_type: selectedTemplate.type,
        data_mapping: mapping
      };
      await apiPost("/widgets", payload);
      success = `Widget "${widget.name}" created successfully.`;
      // Reset form
      widget = { name: "", api_url: "", api_headers: {}, description: "", timeout: 30, enabled: true };
      mapping = {};
      selectedTemplate = null;
    } catch (e) {
      error = "Failed to save widget: " + e.message;
    } finally {
      loading = false;
    }
  }

  onMount(loadTemplates);
</script>

<div class="grid" style="gap: 16px;">
  <div class="grid" style="grid-template-columns: 1fr 1fr; gap: 16px;">
    <div>
      <label class="label">Widget Name *</label>
      <input class="input" bind:value={widget.name} placeholder="e.g., Weather Widget"/>
    </div>
    <div>
      <label class="label">Timeout (seconds)</label>
      <input class="input" type="number" min="5" max="120" bind:value={widget.timeout}/>
    </div>
    <div style="grid-column: 1 / -1;">
      <label class="label">API URL *</label>
      <input class="input" type="url" bind:value={widget.api_url} placeholder="https://api.example.com/data"/>
    </div>
    <div style="grid-column: 1 / -1;">
      <label class="label">Description</label>
      <textarea class="textarea" rows="3" bind:value={widget.description} placeholder="What does this widget display?"></textarea>
    </div>
  </div>

  <div>
    <h3 style="margin: 8px 0;">Choose Template</h3>
    <div class="grid" style="grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));">
      {#each templates as t}
        <button class="btn secondary" class:selected={selectedTemplate?.type === t.type} onclick={() => selectedTemplate = t} style="text-align:left;">
          <div style="font-weight:600;">{t.name}</div>
          <div style="color:#666; font-size:.9rem; margin-top:4px;">{t.description}</div>
        </button>
      {/each}
    </div>
  </div>

  {#if selectedTemplate}
    <div>
      <h3 style="margin: 8px 0;">Field Mapping</h3>
      <div class="grid">
        {#each selectedTemplate.fields || [] as f}
          <div class="grid" style="grid-template-columns: 180px 1fr; align-items:center;">
            <div class="label">{f.label}</div>
            <input class="input" placeholder={f.placeholder} bind:value={mapping[f.key]} oninput={(e) => mapping = { ...mapping, [f.key]: e.currentTarget.value }} />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if error}<div class="status err">{error}</div>{/if}
  {#if success}<div class="status ok">{success}</div>{/if}

  <div style="display:flex; gap: 10px;">
    <button class="btn primary" disabled={loading} onclick={save}>{loading ? "Saving…" : "💾 Save Widget"}</button>
    <button class="btn secondary" onclick={() => { widget = { name:"", api_url:"", api_headers:{}, description:"", timeout:30, enabled:true }; mapping = {}; selectedTemplate=null; }}>Reset</button>
  </div>
</div>
