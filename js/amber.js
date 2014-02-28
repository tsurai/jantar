$('form').on('submit', function(e) {
  var csrf_token = $('meta[name=csrf-token]').attr('content');
  
  if(csrf_token != null) {
    $(this).append('<input type="hidden" name="_csrf-token" value="'+csrf_token+'">');
  }
})

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
})