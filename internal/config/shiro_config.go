package config

// FilterChainDefinition mirrors ShiroConfig#shiroFilterChainDefinition:
//
//	filterMap.put("/**", "jwt");
//
// Every path goes through the jwt filter. The filter itself decides whether an
// anonymous request may proceed, which is why unauthenticated reads still work.
var FilterChainDefinition = map[string]string{
	"/**": "jwt",
}
