# Introduction

This tool exposes database segment content using fuse and allows modification of content using files. This enables core configurations (such as YAML format) and dynamic scripts to be stored in the database.

This tool is suitable for small-scale scenarios, where only one database cluster needs to be maintained without the need for object storage.

With this tool, modifying the corresponding configurations and scripts can be much more comfortable.

# References

Previously, I considered using https://github.com/hanwen/go-fuse, but it was not user-friendly. Later, I referred to the implementation of https://github.com/jacobsa/fuse/samples/hellofs. Thanks!

# Other

The models based on GORM in the code are examples, not real production models. For real scenarios, modifications can be made based on these codes.

