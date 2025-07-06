export interface AuditEvent {
    timestamp: string;
    username: string;
    namespace: string;
    app: string;
    action: string;
    details?: any;
}

export const getAuditEvents = async (): Promise<AuditEvent[]> => {
    const response = await fetch('/api/v1/audit.trail');
    return response.json();
}

export const addAuditEvent = async (event: AuditEvent) => {
    const response = await fetch('/api/v1/audit.trail', {
        method: 'POST',
        body: JSON.stringify(event),
    });

    if (!response.ok) {
        console.error('Failed to add audit event', response.status, await response.json());
    }
}