package domain

// Supplier lifecycle is deliberately small: Active is the operational flag;
// hard deletion is a separate storage operation and never masquerades as archive.
