import{$ as p}from"./dataTables.bootstrap4-2CiN1ToA.js";import{r as g,l as h,a as y,h as m,s as E,e as v,c as b,d as f}from"./main-CfrVZou3.js";p(document).ready(function(){p(".table").DataTable({responsive:{details:{responsive:!0,type:"none",target:""}},order:[[0,"desc"]],language:{search:""}}),p(".dataTables_filter input").attr("placeholder","Search...")});g(function(){const o=document.getElementById("addConfigModal"),e=document.getElementById("add-form");o&&o.addEventListener("show.coreui.modal",d=>{const i=d.relatedTarget.getAttribute("data-id");e.action=`/endpoints/${i}/configs`})});g(function(){const o=document.getElementById("updateConfigModal"),e=document.getElementById("update-form"),d=document.getElementById("update-btn"),r=document.getElementById("update-restart-btn");o&&o.addEventListener("show.coreui.modal",c=>{const t=c.relatedTarget,a=t.getAttribute("data-id");e.querySelector('input[name="endpoint_id"]').value=a;const u=t.getAttribute("data-config-value"),l=t.getAttribute("data-custom")==="true",s=e.querySelector('select[name="config_value"]'),n=e.querySelector('input[name="custom_value"]');l?(s.value="custom",n.value=u,n.classList.remove("d-none")):(s.value=u,n.classList.add("d-none"))});function i(c,t=!1){c.disabled=!0,h();const u=`/endpoints/${e.querySelector('input[name="endpoint_id"]').value}/configs`,l=new FormData(e);l.delete("endpoint_id"),t?l.append("restart",1):l.append("restart",0),e.querySelector('input[name="endpoint_id"]').value="",y(u,{method:"PATCH",redirect:"error",body:l}).then(s=>s.json().then(n=>({ok:s.ok,data:n}))).then(({ok:s,data:n})=>{if(m(),!s)throw new Error(n.error||"Error occurred");n.success?(E(n.success+". Page will refresh shortly..."),setTimeout(()=>{v()},3e3)):n.redirect&&b(n.redirect),c.disabled=!1}).catch(s=>{m(),s.message==="Failed to fetch"?f("There is a problem with processing the request. Reload and try again."):f(s.message),c.disabled=!1})}d.addEventListener("click",function(){i(d)}),r.addEventListener("click",function(){i(r,!0)})});g(function(){const o=document.getElementById("deleteConfigModal"),e=document.getElementById("delete-form"),d=document.getElementById("delete-btn");o&&o.addEventListener("show.coreui.modal",r=>{const i=r.relatedTarget,c=i.getAttribute("data-id");e.querySelector('input[name="endpoint_id"]').value=c,document.getElementById("delete-endpoint-name").textContent=i.getAttribute("data-name")}),d.addEventListener("click",function(){d.disabled=!0,h();const r=e.querySelector('input[name="endpoint_id"]').value,i=e.querySelector('input[name="config_key"]').value,c=`/endpoints/${r}/configs/${i}`;e.querySelector('input[name="endpoint_id"]').value="",document.getElementById("delete-endpoint-name").textContent="",y(c,{method:"DELETE",redirect:"error"}).then(t=>t.json().then(a=>({ok:t.ok,data:a}))).then(({ok:t,data:a})=>{if(m(),!t)throw new Error(a.error||"Error occurred");a.success?(E(a.success+". Page will refresh shortly..."),setTimeout(()=>{v()},3e3)):a.redirect&&b(a.redirect),d.disabled=!1}).catch(t=>{m(),t.message==="Failed to fetch"?f("There is a problem with processing the request. Reload and try again."):f(t.message),d.disabled=!1})})});g(function(){function o(e){const r=e.closest("form").querySelector('input[name="custom_value"]');e.addEventListener("change",function(){this.value==="custom"?r.classList.remove("d-none"):r.classList.add("d-none")})}document.querySelectorAll('form select[name="config_value"]').forEach(function(e){o(e)})});
