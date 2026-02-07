package com.coooolfan.boxlites.core.port

import com.coooolfan.boxlites.core.model.ExecResult

interface BoxSession : AutoCloseable {
    val sessionId: String

    fun run(code: String): ExecResult

    override fun close()
}
