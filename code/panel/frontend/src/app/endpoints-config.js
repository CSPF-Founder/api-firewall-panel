import {
  redirectToLogin,
  showError,
  showSuccess,
  resetInputForm,
  loadingBox,
  hideLoadingBox,
  requestWithCSRFToken,
  ready,
  refreshPage,
} from "./main.js";

import "bootstrap";
import "datatables.net";
import "datatables.net-bs4";

// Only Jquery Dependency
$(document).ready(function () {
  $(".table").DataTable({
    responsive: {
      details: {
        responsive: true,
        type: "none",
        target: "",
      },
    },
    order: [[0, "desc"]],
    language: {
      search: "",
    },
  });

  $(".dataTables_filter input").attr("placeholder", "Search...");
});

// add Endpoint Cofig
ready(function () {
  const addConfigModal = document.getElementById("addConfigModal");
  const addForm = document.getElementById("add-form");

  if (addConfigModal) {
    addConfigModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const endPointID = button.getAttribute("data-id");
      addForm.action = `/endpoints/${endPointID}/configs`;
    });
  }
});

// Update Endpoint Config
ready(function () {
  const updateModal = document.getElementById("updateConfigModal");
  const updateForm = document.getElementById("update-form");
  const updateBtn = document.getElementById("update-btn");
  const updateRestartBtn = document.getElementById("update-restart-btn");

  if (updateModal) {
    updateModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const targetID = button.getAttribute("data-id");
      updateForm.querySelector('input[name="endpoint_id"]').value = targetID;

      const configValue = button.getAttribute("data-config-value");
      const isCustom = button.getAttribute("data-custom") === "true"; // Check if config value is custom
      const selectConfigValue = updateForm.querySelector(
        'select[name="config_value"]'
      );
      const customInput = updateForm.querySelector(
        'input[name="custom_value"]'
      );

      if (isCustom) {
        selectConfigValue.value = "custom";
        customInput.value = configValue;
        customInput.classList.remove("d-none");
      } else {
        selectConfigValue.value = configValue;
        customInput.classList.add("d-none");
      }
    });
  }

  function handleUpdateConfig(clickedBtn, restart = false) {
    clickedBtn.disabled = true;
    loadingBox();

    const endPointID = updateForm.querySelector(
      'input[name="endpoint_id"]'
    ).value;
    const targetURL = `/endpoints/${endPointID}/configs`;

    const formData = new FormData(updateForm);
    //remove the endpoint_id
    formData.delete("endpoint_id");
    if (restart) {
      formData.append("restart", 1);
    } else {
      formData.append("restart", 0);
    }

    // clear the form
    updateForm.querySelector('input[name="endpoint_id"]').value = "";

    requestWithCSRFToken(targetURL, {
      method: "PATCH",
      redirect: "error",
      body: formData,
    })
      .then((response) =>
        response.json().then((data) => ({ ok: response.ok, data }))
      )
      .then(({ ok, data }) => {
        hideLoadingBox();
        if (!ok) {
          throw new Error(data.error || "Error occurred");
        }
        if (data.success) {
          showSuccess(data.success + ". Page will refresh shortly...");
          setTimeout(() => {
            refreshPage();
          }, 3000);
        } else if (data.redirect) {
          redirectToLogin(data.redirect);
        }
        clickedBtn.disabled = false;
      })
      .catch((error) => {
        hideLoadingBox();
        if (error.message === "Failed to fetch") {
          showError(
            "There is a problem with processing the request. Reload and try again."
          );
        } else {
          showError(error.message);
        }
        clickedBtn.disabled = false;
      });
  }

  updateBtn.addEventListener("click", function () {
    handleUpdateConfig(updateBtn);
  });

  updateRestartBtn.addEventListener("click", function () {
    handleUpdateConfig(updateRestartBtn, true);
  });
});

// Delete Endpoint Config
ready(function () {
  const deleteModal = document.getElementById("deleteConfigModal");
  const deleteForm = document.getElementById("delete-form");
  const deleteEndpointBtn = document.getElementById("delete-btn");

  if (deleteModal) {
    deleteModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const targetID = button.getAttribute("data-id");
      deleteForm.querySelector('input[name="endpoint_id"]').value = targetID;
      document.getElementById("delete-endpoint-name").textContent =
        button.getAttribute("data-name");
    });
  }

  deleteEndpointBtn.addEventListener("click", function () {
    deleteEndpointBtn.disabled = true;
    loadingBox();

    const targetID = deleteForm.querySelector(
      'input[name="endpoint_id"]'
    ).value;

    const configKey = deleteForm.querySelector(
      'input[name="config_key"]'
    ).value;

    const deleteEndpointURL = `/endpoints/${targetID}/configs/${configKey}`;

    // clear the form
    deleteForm.querySelector('input[name="endpoint_id"]').value = "";
    document.getElementById("delete-endpoint-name").textContent = "";

    requestWithCSRFToken(deleteEndpointURL, {
      method: "DELETE",
      redirect: "error",
    })
      .then((response) =>
        response.json().then((data) => ({ ok: response.ok, data }))
      )
      .then(({ ok, data }) => {
        hideLoadingBox();
        if (!ok) {
          throw new Error(data.error || "Error occurred");
        }
        if (data.success) {
          showSuccess(data.success + ". Page will refresh shortly...");
          setTimeout(() => {
            refreshPage();
          }, 3000);
        } else if (data.redirect) {
          redirectToLogin(data.redirect);
        }
        deleteEndpointBtn.disabled = false;
      })
      .catch((error) => {
        hideLoadingBox();
        if (error.message === "Failed to fetch") {
          showError(
            "There is a problem with processing the request. Reload and try again."
          );
        } else {
          showError(error.message);
        }
        deleteEndpointBtn.disabled = false;
      });
  });
});

ready(function () {
  function handleConfigValueChangeWithinForm(selectElement) {
    const form = selectElement.closest("form");
    const customInput = form.querySelector('input[name="custom_value"]');

    selectElement.addEventListener("change", function () {
      if (this.value === "custom") {
        customInput.classList.remove("d-none");
      } else {
        customInput.classList.add("d-none");
      }
    });
  }

  document
    .querySelectorAll('form select[name="config_value"]')
    .forEach(function (selectElement) {
      handleConfigValueChangeWithinForm(selectElement);
    });
});
