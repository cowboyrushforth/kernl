
record_like = (news_id, done) ->
  $.ajax
    url: "/news/like"
    dataType: 'json'
    data:
      news_id: news_id
    success: (result) ->
      done result

$ ->
  $('.news-like').click (e) ->
    e.preventDefault()
    summary = $(e.target).parent().parent().find('.likes-summary')
    record_like 123, (result) ->
      summary.show()

  $('.news-comment').click (e) ->
    e.preventDefault()

  $('.news-share').click (e) ->
    e.preventDefault()

  $('.comment-like').click (e) ->
    e.preventDefault()
    likes = $(e.target).parent().find('.display-likes')
    likes.show()

  $('div#ghetto-post textarea').focus (e) ->
    e.preventDefault()
    $(this).parent().parent().removeClass('closed')
    $(this).parent().parent().find('div.button-row').show()

  $('div#ghetto-post textarea').blur (e) ->
    e.preventDefault()
    v = $.trim($(this).val())
    $(this).parent().parent().addClass('closed') if v == ""
    $(this).parent().parent().find('div.button-row').hide() if v == ""

  $('button[data-dismiss="alert"]').click (e) ->
    e.preventDefault()
    $(this).parent().remove()
