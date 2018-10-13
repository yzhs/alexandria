# Alexandria

Alexandria brings mathematical knowledge to your fingertips.  It provides
full-text search for LaTeX documents.

## What problem does Alexandria try to solve?
When working as a mathematician you will frequently be facing the problem of
finding theorems or definitions that apply to the problem at hand.  Sure, for
commonly used theorems, Google will be all you will need.  For more specialized
knowledge, you can use use some full-text search solution for PDFs, so you can
search a bunch of books or papers.

When you search for something containing the terms "closure" and "compact", a
PDF based search might show you a page with those terms in close proximity.
The matches will surely include the theorem you are looking for, but it could
just as well be a page containing two theorems, one ending with "Hence, A is
compact." and the one after that starting with "Let X be the closure of …".

Generally, you will have lots of false positive matches, where the search terms
occur close together, but *not* as part of the same theorem.  The problem, of
course, is that PDF documents lack structure.  A program reading a PDF document
has no way of knowing where one theorem ends and the next one begins.  A
full-text search can only look for proximity, not for *is part of the same
theorem.*

## How is Alexandria different?
Alexandria does nothing more than a full-text search of some documents.  Its
advantage comes from the ability to have one document per theorem or definition.
That way, if a match in a document is found, you have, by definition, found a
theorem or definition that contains the search terms.

Each document has a type, and can contain tags.  In addition, you can search by
type, e.g. using `type:definition`, or tag, e.g. `tag:topology tag:analysisc`.
You can also add a source annotation, which can also be searched.

## How to interact with Alexandria
You can interact with Alexandria in two different ways:

* A command line interface, also called `alexandria`

  The CLI can be used to update the index, using `alexandria -i` or `--index`,
  print statistics using `-S` or `--stats`, or to query the Alexandria knowledge
  base by just calling `alexandria` followed by some terms you want to search
  for.  Note that you might have to escape some characters, depending on your
  shell and its configuration.

* A web interface, `alexandria-web`.

  `alexandria-web` start a web server that listens on `127.0.0.1:41665`.  Visit
  that page with a web browser of your choice to use `alexandria-web`.

### Search
Say you want to look up some definition from Hartshorne's *Algebraic Geometry*
mentioning the Zariski topology.  You could search for `source:hartshorne
tag:geometry type:definition zariski`.
