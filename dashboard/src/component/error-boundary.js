"use client";

import { ErrorBoundary as EB } from "react-error-boundary";

function fallbackRender({ error, resetErrorBoundary }) {
    // Call resetErrorBoundary() to reset the error boundary and retry the render.
    return (
        <div role="alert">
            <pre style={{ color: "red" }}>{error.message}</pre>
        </div>
    );
}


export default function ErrorBoundary({ key, children }) {
    return (
        <EB
            fallbackRender={fallbackRender}
        >
            {children}
        </EB>
    )
}


export function withErrorBoundary(Component) {
    return function (props) {
        return (
            <EB
                fallbackRender={fallbackRender}
            >
                <Component {...props} />
            </EB>
        )
    }
}
