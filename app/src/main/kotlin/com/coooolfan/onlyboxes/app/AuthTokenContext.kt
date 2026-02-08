package com.coooolfan.onlyboxes.app

import com.coooolfan.onlyboxes.core.exception.CodeExecutionException

interface AuthTokenProvider {
    fun requireToken(): String
}

class ThreadLocalAuthTokenProvider : AuthTokenProvider {
    override fun requireToken(): String {
        val token = AuthTokenContextHolder.currentToken()?.trim()
        if (token.isNullOrEmpty()) {
            throw CodeExecutionException("Authentication token context is missing")
        }
        return token
    }
}

internal object AuthTokenContextHolder {
    private val tokenContext = ThreadLocal<String?>()

    fun bindToken(token: String) {
        tokenContext.set(token)
    }

    fun currentToken(): String? = tokenContext.get()

    fun clear() {
        tokenContext.remove()
    }
}
