- Add test for handling of template parameters containing imported
  types (like list.List) in rungo.go.

- In general, create a decent test suite, which can also serve as a
  sample template library.

- Make gotimports search the GOROOT for templates.
