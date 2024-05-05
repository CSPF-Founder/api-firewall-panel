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

// add Endpoint
ready(function () {
  const addForm = document.getElementById("add-endpoint-form");
  if (!addForm) {
    return;
  }

  const addButton = document.getElementById("add-endpoint-btn");
  const addEndpointURL = addForm.getAttribute("action");
  // Add Scan Form Submit

  addForm.addEventListener("submit", function (event) {
    addButton.disabled = true;

    event.preventDefault();
    loadingBox();

    const formData = new FormData(addForm);

    requestWithCSRFToken(addEndpointURL, {
      method: "POST",
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
          $("#port_input").addClass("d-none");
          resetInputForm("#add-endpoint-form");
          showSuccess(data.success);
        } else if (data.redirect) {
          redirectToLogin(data.redirect);
        }

        addButton.disabled = false;
      })
      .catch((error) => {
        hideLoadingBox();
        showError(error.message);
        addButton.disabled = false;
      });
  });
});

// update Endpoint
ready(function () {
  const editForm = document.getElementById("edit-endpoint-form");
  if (!editForm) {
    return;
  }

  const addButton = document.getElementById("edit-endpoint-btn");
  const apiUrl = document.getElementById("api-url");
  const request_mode = document.getElementById("request-mode");
  const editEndpointURL = editForm.getAttribute("action");

  // Edit Endpoint Form Submit
  editForm.addEventListener("submit", function (event) {
    addButton.disabled = true;

    event.preventDefault();
    loadingBox();

    const formData = new FormData(editForm);

    requestWithCSRFToken(editEndpointURL, {
      method: "PATCH",
      body: formData,
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
          resetInputForm("#edit-endpoint-form");
          request_mode.value = data.request_mode;
          apiUrl.value = data.api_url;
          showSuccess(data.success);
        } else if (data.redirect) {
          redirectToLogin(data.redirect);
        }

        addButton.disabled = false;
      })
      .catch((error) => {
        hideLoadingBox();
        showError(error.message);
        addButton.disabled = false;
      });
  });
});

// Change Endpoint Request Mode
ready(function () {
  const requestUpdateModal = document.getElementById("requestUpdateModal");
  const endPointTable = document.getElementById("endpoint-table");
  const updateModeForm = document.getElementById("update-request-mode-form");
  const updateModeBtn = document.getElementById("update-request-mode-btn");

  if (!endPointTable) {
    return;
  }

  if (requestUpdateModal) {
    requestUpdateModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const entryID = button.getAttribute("data-id");
      const requestMode = button.getAttribute("data-request");
      updateModeForm.querySelector('select[name="request_mode"]').value =
        requestMode;
      updateModeForm.querySelector('input[name="endpoint_id"]').value = entryID;
    });
  }

  updateModeBtn.addEventListener("click", function () {
    updateModeBtn.disabled = true;
    loadingBox();

    const entryID = updateModeForm.querySelector(
      'input[name="endpoint_id"]'
    ).value;
    const updateURL = `/endpoints/${entryID}/mode`;

    const formData = new FormData(updateModeForm);
    //remove the endpoint_id
    formData.delete("endpoint_id");

    // clear the form
    updateModeForm.querySelector('select[name="request_mode"]').value = "";
    updateModeForm.querySelector('input[name="endpoint_id"]').value = "";

    requestWithCSRFToken(updateURL, {
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
        updateModeBtn.disabled = false;
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
        updateModeBtn.disabled = false;
      });
  });
});

// Delete Endpoint
ready(function () {
  const deleteModal = document.getElementById("deleteModal");
  const endPointTable = document.getElementById("endpoint-table");
  const deleteForm = document.getElementById("delete-endpoint-form");
  const deleteEndpointBtn = document.getElementById("delete-endpoint-btn");

  if (!endPointTable) {
    return;
  }

  if (deleteModal) {
    deleteModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const entryID = button.getAttribute("data-id");
      deleteForm.querySelector('input[name="endpoint_id"]').value = entryID;
      document.getElementById("delete-endpoint-name").textContent =
        button.getAttribute("data-name");
    });
  }

  deleteEndpointBtn.addEventListener("click", function () {
    deleteEndpointBtn.disabled = true;
    loadingBox();

    const entryID = deleteForm.querySelector('input[name="endpoint_id"]').value;

    const deleteEndpointURL = `/endpoints/${entryID}`;

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
          const rowToRemove = document.querySelector(
            `#endpoint-table tr[data-id="${entryID}"]`
          );
          if (rowToRemove) {
            rowToRemove.parentNode.removeChild(rowToRemove);
          }

          showSuccess(data.success);
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

// Restart Endpoint
ready(function () {
  const restartModal = document.getElementById("restartModal");
  const endPointTable = document.getElementById("endpoint-table");
  const restartForm = document.getElementById("restart-form");

  if (!endPointTable) {
    return;
  }

  if (restartModal) {
    restartModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const entryID = button.getAttribute("data-id");
      document.getElementById("restart-endpoint-name").textContent =
        button.getAttribute("data-name");

      restartForm.setAttribute("action", `/endpoints/${entryID}/restart`);
    });
  }
});

//Change port input
ready(function () {
  var select = document.getElementById("port_select");
  if (!select) {
    return;
  }
  select.addEventListener("change", function () {
    var val = this.value;
    let input = document.getElementById("port_input");
    console.log(val);
    if (val == "auto") {
      input.classList.add("d-none");
    } else {
      input.classList.remove("d-none");
    }
  });
});
