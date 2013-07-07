
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
    record_like 123, (result) ->
      console.log result

