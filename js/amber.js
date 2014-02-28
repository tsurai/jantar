$('form').on('submit', function(e) {
  var csrf_token = $('meta[name=csrf-token]').attr('content');
  
  if(csrf_token != null) {
    $(this).append('<input type="hidden" name="_csrf-token" value="'+csrf_token+'">');
  }
});

$('a[data-method]').on('click', function(e) {
  var csrf_token = $('meta[name=csrf-token]').attr('content');
  var href = $(this).attr('href');
  var method = $(this).attr('data-method').toUpperCase();

  if(!(method == "GET" || method == "POST" || method == "PUT" || method == "GET")) {
    method = "POST";
  }
  
  var form = $('<form method="POST" href="'+href+'">');
  form.append('<input type="hidden" name="_method" value="'+method+'"/>')
      .append('<input type="hidden" name="_csrf-token" value="'+csrf_token+'"/>')
      .appendTo('body');

  e.preventDefault();
  form.submit();
});

$('a[data-confirm], input[data-confirm]').on('click', function() {
  if(!confirm($(this).attr('data-confirm'))) {
    return false;
  }
});

$('a[data-confirm]').on('click', function() {
  var href = $(this).attr('href');
  if (!$('#dataConfirmModal').length) {
    $('body').append('<div id="dataConfirmModal" class="modal fade bs-modal-sm" tabindex="-1" role="dialog" aria-labelledby="mySmallModalLabel" aria-hidden="true"><div class="modal-dialog modal-sm"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button><h3 id="dataConfirmLabel">Please Confirm</h3></div><div class="modal-body"></div><div class="modal-footer"><button class="btn" data-dismiss="modal" aria-hidden="true">Cancel</button><a class="btn btn-primary" id="dataConfirmOK">OK</a></div></div></div></div>');
  } 
  
  $('#dataConfirmModal').find('.modal-body').text($(this).attr('data-confirm'));
  $('#dataConfirmOK').attr('href', href);
  $('#dataConfirmModal').modal({show:true});
  
  return false;
});