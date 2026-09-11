package config

// PaginationEnabled mirrors MybatisPlusConfig#paginationInterceptor.
//
// Registering PaginationInterceptor is what makes Page actually apply LIMIT and
// run the COUNT query; without it MP would return every row. The Go mappers
// implement LIMIT/COUNT directly, so this constant documents that the behaviour
// is intentionally on rather than gating anything.
const PaginationEnabled = true

// TransactionManagementEnabled mirrors @EnableTransactionManagement. vueblog
// declares no @Transactional methods, so this changes no observable behaviour.
const TransactionManagementEnabled = true
