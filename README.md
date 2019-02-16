This is a small helper package for building and tearing down resources used by micro-services. The library models these resources with a generic `Resource` data structure.

Once resources define an ordered structure, they can be created/ torn down using that structure. The hope is that developers using this package may modularize thier code. 

One such scheme is to fashion resource builders in separate packages, with pre-defined `[]Property` dependencies.
