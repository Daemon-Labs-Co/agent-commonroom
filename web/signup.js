(function () {
  const $ = (id) => document.getElementById(id);
  const form = $("signup-form"), nameInput = $("signup-name"), idInput = $("signup-id"), codeInput = $("signup-code");
  let idEdited = false;

  // The address follows the name until the person edits it themselves.
  function slug(text) {
    return String(text).toLowerCase().normalize("NFKD").replace(/[̀-ͯ]/g, "").replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 63);
  }
  nameInput.addEventListener("input", () => { if (!idEdited) idInput.value = slug(nameInput.value); });
  idInput.addEventListener("input", () => { idEdited = true; });

  function fail(message) {
    const box = $("signup-error");
    box.textContent = message;
    box.hidden = !message;
  }

  fetch("/api/signup/config").then((r) => r.ok ? r.json() : null).then((config) => {
    if (!config) return;
    $("signup-domain").textContent = "." + config.domain;
    $("signup-code-field").hidden = !config.codeRequired;
  }).catch(() => {});

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    fail("");
    const name = nameInput.value.trim(), id = idInput.value.trim();
    if (!name || !id) { fail("Enter a workspace name and an address."); return; }
    const button = $("signup-submit");
    button.disabled = true;
    try {
      const response = await fetch("/api/signup", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name, id, code: codeInput.value }) });
      const text = await response.text();
      if (!response.ok) { fail(text.trim() || "Sign-up failed."); button.disabled = false; return; }
      const result = JSON.parse(text);
      const link = $("signup-link");
      link.href = result.setupLink;
      link.textContent = result.setupLink;
      $("signup-expiry").textContent = "The link expires " + new Date(result.setupExpires).toLocaleString() + ". Keep it private.";
      form.hidden = true;
      $("signup-done").hidden = false;
    } catch (e) {
      fail("Sign-up could not be reached. Try again.");
      button.disabled = false;
    }
  });
})();
