# REST boundary

REST handlers belong at the edge and translate HTTP requests into the internal protobuf/domain contracts. Authentication, validation, rate limits and trace propagation must occur before invoking core operations.
