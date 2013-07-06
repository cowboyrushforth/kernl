class Newsfeed
  constructor: (div) ->
    @div = div
    @items = [{name: 'Fred Cline', message: 'hey I love coffee'},
              {name: 'Bernard Tomic', message: 'hey I love tennis'}]
  render: ->
    console.log @items


nf = new Newsfeed(null)
console.log nf.render()
