package com.coooolfan.onlyboxes.app

import jakarta.servlet.FilterChain
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Value
import org.springframework.boot.web.servlet.FilterRegistrationBean
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.core.Ordered
import org.springframework.web.filter.OncePerRequestFilter
import java.nio.charset.StandardCharsets

private val LOWER_ALNUM_TOKEN_REGEX = Regex("^[a-z0-9]+$")
private const val DEFAULT_AUTH_HEADER_NAME = "X-Onlyboxes-Token"
private const val DEFAULT_MCP_ENDPOINT = "/mcp"
private const val UNAUTHORIZED_BODY = "Unauthorized"
private const val UNAUTHORIZED_CONTENT_TYPE = "text/plain;charset=UTF-8"

@Configuration(proxyBeanMethods = false)
class AuthConfig {

    private val log = LoggerFactory.getLogger(AuthConfig::class.java)

    @Bean
    fun tokenAuthFilterRegistration(
        @Value("\${onlyboxes.auth.header-name:$DEFAULT_AUTH_HEADER_NAME}")
        headerNameRaw: String,
        @Value("\${onlyboxes.auth.allowed-tokens:}")
        allowedTokensRaw: String,
        @Value("\${spring.ai.mcp.server.streamable-http.mcp-endpoint:$DEFAULT_MCP_ENDPOINT}")
        mcpEndpointRaw: String,
    ): FilterRegistrationBean<TokenAuthFilter> {
        val headerName = parseHeaderName(headerNameRaw)
        val allowedTokens = parseAllowedTokens(allowedTokensRaw)
        val mcpEndpoint = normalizeMcpEndpoint(mcpEndpointRaw)

        log.info("Auth Config: load ${allowedTokens.size} tokens for header '$headerName' on endpoint '$mcpEndpoint'")

        return FilterRegistrationBean<TokenAuthFilter>().apply {
            setFilter(
                TokenAuthFilter(
                    headerName = headerName,
                    allowedTokens = allowedTokens,
                    protectedEndpoint = mcpEndpoint,
                ),
            )
            setName("tokenAuthFilter")
            order = Ordered.HIGHEST_PRECEDENCE
            addUrlPatterns(*resolveUrlPatterns(mcpEndpoint).toTypedArray())
        }
    }
}

class TokenAuthFilter(
    private val headerName: String,
    private val allowedTokens: Set<String>,
    private val protectedEndpoint: String,
) : OncePerRequestFilter() {
    override fun shouldNotFilter(request: HttpServletRequest): Boolean {
        return !matchesProtectedEndpoint(request)
    }

    override fun doFilterInternal(
        request: HttpServletRequest,
        response: HttpServletResponse,
        filterChain: FilterChain,
    ) {
        val token = request.getHeader(headerName)
        if (token == null || !LOWER_ALNUM_TOKEN_REGEX.matches(token) || !allowedTokens.contains(token)) {
            rejectUnauthorized(response)
            return
        }
        filterChain.doFilter(request, response)
    }

    private fun matchesProtectedEndpoint(request: HttpServletRequest): Boolean {
        if (protectedEndpoint == "/") {
            return true
        }
        val contextPath = request.contextPath.orEmpty()
        val path = request.requestURI.removePrefix(contextPath)
        return path == protectedEndpoint || path.startsWith("$protectedEndpoint/")
    }

    private fun rejectUnauthorized(response: HttpServletResponse) {
        response.status = HttpServletResponse.SC_UNAUTHORIZED
        response.characterEncoding = StandardCharsets.UTF_8.name()
        response.contentType = UNAUTHORIZED_CONTENT_TYPE
        response.writer.write(UNAUTHORIZED_BODY)
    }
}

internal fun parseAllowedTokens(allowedTokensRaw: String): Set<String> {
    val tokens = linkedSetOf<String>()
    allowedTokensRaw
        .split(",")
        .map { it.trim() }
        .filter { it.isNotEmpty() }
        .forEach { token ->
            if (!LOWER_ALNUM_TOKEN_REGEX.matches(token)) {
                throw IllegalStateException(
                    "onlyboxes.auth.allowed-tokens contains invalid token: tokens must match ^[a-z0-9]+$",
                )
            }
            tokens.add(token)
        }
    return tokens
}

internal fun parseHeaderName(headerNameRaw: String): String {
    val headerName = headerNameRaw.trim()
    if (headerName.isEmpty()) {
        throw IllegalStateException("onlyboxes.auth.header-name must not be blank")
    }
    return headerName
}

internal fun normalizeMcpEndpoint(mcpEndpointRaw: String): String {
    val endpoint = mcpEndpointRaw.trim()
    if (endpoint.isEmpty()) {
        throw IllegalStateException("spring.ai.mcp.server.streamable-http.mcp-endpoint must not be blank")
    }
    val withLeadingSlash = if (endpoint.startsWith("/")) endpoint else "/$endpoint"
    return withLeadingSlash.trimEnd('/').ifEmpty { "/" }
}

private fun resolveUrlPatterns(mcpEndpoint: String): List<String> {
    return if (mcpEndpoint == "/") {
        listOf("/*")
    } else {
        listOf(mcpEndpoint, "$mcpEndpoint/*")
    }
}
