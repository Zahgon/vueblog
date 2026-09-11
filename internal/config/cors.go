package config

import "net/http"

// Cors mirrors com.markerhub.config.CorsConfig#addCorsMappings.
//
//	registry.addMapping("/**")
//	        .allowedOrigins("*")
//	        .allowedMethods("GET","HEAD","POST","PUT","DELETE","OPTIONS")
//	        .allowCredentials(true)
//	        .maxAge(3600)
//	        .allowedHeaders("*");
//
// Spring resolves allowedOrigins("*") with allowCredentials(true) by echoing
// the request Origin rather than literally sending "*", because browsers reject
// the wildcard on credentialed requests. That echoing behaviour is reproduced.
func Cors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")
	if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
		w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}
}
