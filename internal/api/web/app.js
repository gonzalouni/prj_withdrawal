"use strict";
const $ = (id) => document.getElementById(id);
const sel = $("cuenta"), saldo = $("saldo"), msg = $("msg"), tbody = $("historial");

function soles(n) { return "S/ " + Number(n).toFixed(2); }

function mostrar(clase, texto) { msg.className = clase; msg.textContent = texto; }

async function api(ruta, opciones) {
  const r = await fetch(ruta, opciones);
  const cuerpo = await r.json();
  return { ok: r.ok, status: r.status, cuerpo };
}

async function cargarCuentas() {
  const { cuerpo } = await api("/api/cuentas");
  sel.replaceChildren(...cuerpo.map((c) => new Option(c.nombre, c.nombre)));
  await refrescar();
}

async function refrescar() {
  const nombre = sel.value;
  const [c, h] = await Promise.all([
    api("/api/cuentas/" + encodeURIComponent(nombre)),
    api("/api/transacciones?cuenta=" + encodeURIComponent(nombre)),
  ]);
  saldo.textContent = soles(c.cuerpo.saldo);
  tbody.replaceChildren(...h.cuerpo.slice().reverse().map(fila));
}

function celda(texto, clase) {
  const td = document.createElement("td");
  td.textContent = texto;
  if (clase) td.className = clase;
  return td;
}

function fila(t) {
  const tr = document.createElement("tr");
  const estado = document.createElement("td");
  const b = document.createElement("span");
  b.className = "badge " + t.estado;
  b.textContent = t.estado;
  estado.appendChild(b);
  tr.append(celda(t.id), celda(t.tipo), celda(soles(t.monto)), estado,
            celda(soles(t.saldo_despues)), celda(t.motivo, "det"));
  return tr;
}

$("form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const boton = e.target.querySelector("button");
  boton.disabled = true;
  try {
    const tipo = e.target.elements.tipo.value;
    const { cuerpo } = await api("/api/operaciones", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ cuenta: sel.value, tipo, monto: $("monto").value.trim() }),
    });
    if (cuerpo.estado === "EXITOSA") {
      mostrar("ok", `${cuerpo.id}: operación exitosa. Nuevo saldo ${soles(cuerpo.saldo_despues)}.`);
      $("monto").value = "";
    } else if (cuerpo.estado === "RECHAZADA") {
      mostrar("warn", `${cuerpo.id}: ${cuerpo.motivo}.`);
    } else {
      mostrar("err", cuerpo.motivo || cuerpo.error || "Error inesperado.");
    }
    await refrescar();
  } catch (err) {
    mostrar("err", "No se pudo contactar al servidor.");
  } finally {
    boton.disabled = false;
  }
});

sel.addEventListener("change", () => { mostrar("", ""); refrescar(); });
cargarCuentas();
