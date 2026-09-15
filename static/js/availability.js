(() => {
  const input = document.getElementById("uptime_filter_input");
  const output = document.getElementById("uptime_filter_value");

  if (!input || !output) {
    return;
  }

  const updateValue = () => {
    const value = Number(input.value);
    output.textContent = `${value.toFixed(2).replace(".", ",")}%`;
    input.setAttribute("aria-valuetext", output.textContent);
  };

  input.addEventListener("input", updateValue);
  updateValue();
})();
