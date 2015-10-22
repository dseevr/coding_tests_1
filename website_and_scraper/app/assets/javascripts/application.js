// This is a manifest file that'll be compiled into application.js, which will include all the files
// listed below.
//
// Any JavaScript/Coffee file within this directory, lib/assets/javascripts, vendor/assets/javascripts,
// or any plugin's vendor/assets/javascripts directory can be referenced here using a relative path.
//
// It's not advisable to add code directly here, but if you do, it'll appear at the bottom of the
// compiled file.
//
// Read Sprockets README (https://github.com/rails/sprockets#sprockets-directives) for details
// about supported directives.
//
//= require jquery
//= require jquery_ujs
//= require_tree .

var remove_errors = function() {
  $("#error_explanation").hide();

  $("form input").each(function(i, input) {
      $(input).removeClass("field_with_errors");
  });

  $("form p.error_message").each(function(i, p) {
    p.remove();
  })
}

$(function() {

  $("tbody tr").click(function(e) {
    e.preventDefault();
    var tr = $(this);
    var modal = $("#edit-modal")
    var form = modal.find("form");

    // update the action attribute on the <form> to our JSON endpoint
    form.attr("action", "/listings/" + tr.attr("data-id")  + ".json");

    // and record the tr id
    form.attr("data-tr-id", tr.attr("data-id"));

    // copy data from this <tr> to the modal
    ["global-rank", "name", "url", "description"].forEach(function(field) {
      var new_value = tr.children("td[data-column=" + field +"]").text()
      $("#modal-field-" + field).val(new_value);
    })

    remove_errors();

    modal.modal("show");

    return false;
  })

  $("#modal-submit-button").click(function(e) {
    e.preventDefault();

    var modal = $("#edit-modal");
    var form = modal.find("form");
    var action = form.attr("action");

    // should disable the submit button until the request completes/times out, etc.

    $.ajax({
      type: "POST",
      url: action,
      data: form.serializeArray(),
      error: function(xhr) {
        remove_errors();

        $("#error_explanation").show();

        // add the "field_with_errors" class to any bad fields
        for(var key in xhr.responseJSON) {
          var input = form.find("#modal-field-" + key.replace("_", "-"))
          input.addClass("field_with_errors");

          var error_text = key + " " + xhr.responseJSON[key]
          input.parent().append("<p class=\"error_message\">" + error_text + "</p>");
        }

      },
    }).done(function(data) {
      var tr = $("tr[data-id=" + form.attr("data-tr-id") + "]");

      // copy data from modal back to the <tr>
      ["global-rank", "name", "description"].forEach(function(field) {
        var new_value = $("#modal-field-" + field).val();
        tr.children("td[data-column=" + field +"]").text(new_value);
      });

      // URL has to be done manually...
      var new_value = $("#modal-field-url").val();
      var hyperlink = tr.children("td[data-column=url]").find("a");
      hyperlink.attr("href", new_value);
      hyperlink.text(new_value);

      modal.modal("hide");
    });

  });

});

