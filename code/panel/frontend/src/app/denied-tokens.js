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

// add entry
ready(function () {
  const addForm = document.getElementById("add-form");
  const addBtn = document.getElementById("add-btn");
  const addRestartBtn = document.getElementById("add-restart-btn");

  function handleAdd(clickedBtn, restart = false) {
    clickedBtn.disabled = true;
    loadingBox();

    // based on restart valeue, update input in addform
    if (restart) {
      addForm.querySelector('input[name="restart"]').value = 1;
    } else {
      addForm.querySelector('input[name="restart"]').value = 0;
    }

    addForm.submit();
  }

  addBtn.addEventListener("click", function () {
    handleAdd(addBtn);
  });

  addRestartBtn.addEventListener("click", function () {
    handleAdd(addRestartBtn, true);
  });
});

// Update Denied Token
ready(function () {
  const updateModal = document.getElementById("updateModal");
  const updateForm = document.getElementById("update-form");
  const updateBtn = document.getElementById("update-btn");
  const updateRestartBtn = document.getElementById("update-restart-btn");

  if (updateModal) {
    updateModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;

      const endpointID = button.getAttribute("data-endpoint-id");
      const entryID = button.getAttribute("data-id");
      const entryValue = button.getAttribute("data-entry-value");

      updateForm.querySelector('input[name="endpoint_id"]').value = endpointID;
      updateForm.querySelector('input[name="entry_id"]').value = entryID;
      document.getElementById("updated-denied-token").innerHTML = entryValue;
    });
  }

  function handleUpdate(clickedBtn, restart = false) {
    clickedBtn.disabled = true;
    loadingBox();

    const endpointID = updateForm.querySelector(
      'input[name="endpoint_id"]'
    ).value;

    const entryID = updateForm.querySelector('input[name="entry_id"]').value;

    const updateURL = `/endpoints/${endpointID}/denied-tokens/${entryID}`;

    const formData = new FormData(updateForm);
    //remove the endpoint_id
    formData.delete("endpoint_id");
    if (restart) {
      formData.append("restart", 1);
    } else {
      formData.append("restart", 0);
    }

    // clear the form
    resetInputForm("#update-form");

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
          }, 4000);
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
    handleUpdate(updateBtn);
  });

  updateRestartBtn.addEventListener("click", function () {
    handleUpdate(updateRestartBtn, true);
  });
});

// Delete Denied Token
ready(function () {
  const deleteModal = document.getElementById("deleteModal");
  const deleteForm = document.getElementById("delete-form");
  const deleteBtn = document.getElementById("delete-btn");

  if (deleteModal) {
    deleteModal.addEventListener("show.coreui.modal", (event) => {
      const button = event.relatedTarget;
      const endpointID = button.getAttribute("data-endpoint-id");
      const entryID = button.getAttribute("data-id");

      deleteForm.querySelector('input[name="endpoint_id"]').value = endpointID;
      deleteForm.querySelector('input[name="entry_id"]').value = entryID;

      document.getElementById("delete-endpoint-name").textContent =
        button.getAttribute("data-endpoint-name");
    });
  }

  deleteBtn.addEventListener("click", function () {
    deleteBtn.disabled = true;
    loadingBox();

    const endpointID = deleteForm.querySelector(
      'input[name="endpoint_id"]'
    ).value;

    const entryID = deleteForm.querySelector('input[name="entry_id"]').value;

    const deleteURL = `/endpoints/${endpointID}/denied-tokens/${entryID}`;

    // clear the form
    resetInputForm("#delete-form");

    document.getElementById("delete-endpoint-name").textContent = "";

    requestWithCSRFToken(deleteURL, {
      method: "DELETE",
      redirect: "error",
    })
      .then((response) =>
        response.json().then((data) => ({ ok: response.ok, data }))
      )
      .then(({ ok, data }) => {
        hideLoadingBox();
        // if (!ok) {
        //   throw new Error(data.error || "Error occurred");
        // }
        deleteBtn.disabled = false;

        if (data.is_removed) {
          const rowToRemove = document.querySelector(
            `#denied-tokens-table tr[data-id="${entryID}"]`
          );
          if (rowToRemove) {
            rowToRemove.parentNode.removeChild(rowToRemove);
          }

          if (data.error) {
            showError(data.error);
          } else {
            showSuccess("Entry removed successfully.");
          }
        } else if (data.redirect) {
          redirectToLogin(data.redirect);
        } else {
          if (data.error) {
            showError(data.error);
          } else {
            showError("Failed to remove the entry.");
          }
        }
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
        deleteBtn.disabled = false;
      });
  });
});
