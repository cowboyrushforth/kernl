
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

$ ->
  $('.news-comment').click (e) ->
    e.preventDefault()

$ ->
  $('.news-share').click (e) ->
    e.preventDefault()

$ ->
  $('.comment-like').click (e) ->
    e.preventDefault()
    num_likes = $(e.target).parent().find('.num-likes')
    num_likes.show()
