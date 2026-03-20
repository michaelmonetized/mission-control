/**
 * GitHub Webhook Handler
 * 
 * Receives events from GitHub:
 * - PR opened/commented/approved → creates/updates thread
 * - Issue opened/commented → creates/updates thread
 * 
 * Route: POST /api/github/webhook
 * Expects: X-Hub-Signature-256 header for verification
 */

import crypto from 'crypto';
import { NextRequest, NextResponse } from 'next/server';

const GITHUB_WEBHOOK_SECRET = process.env.GITHUB_WEBHOOK_SECRET || '';

/**
 * Verify GitHub webhook signature
 */
function verifySignature(payload: string, signature: string): boolean {
  if (!GITHUB_WEBHOOK_SECRET) {
    console.warn('[GitHub Webhook] No secret configured. Skipping verification.');
    return true; // Skip verification in dev
  }

  const hash = crypto
    .createHmac('sha256', GITHUB_WEBHOOK_SECRET)
    .update(payload)
    .digest('hex');

  const expectedSignature = `sha256=${hash}`;
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

/**
 * POST /api/github/webhook
 */
export async function POST(request: NextRequest) {
  try {
    // Read raw payload for signature verification
    const payload = await request.text();
    const signature = request.headers.get('x-hub-signature-256') || '';

    // Verify signature
    if (!verifySignature(payload, signature)) {
      console.warn('[GitHub Webhook] Signature verification failed');
      return NextResponse.json(
        { error: 'Invalid signature' },
        { status: 401 }
      );
    }

    // Parse event
    const event = JSON.parse(payload);
    const eventType = request.headers.get('x-github-event');

    console.log(`[GitHub Webhook] Received event: ${eventType}`);

    // Handle different event types
    switch (eventType) {
      case 'pull_request':
        await handlePullRequest(event);
        break;

      case 'pull_request_review':
        await handlePullRequestReview(event);
        break;

      case 'pull_request_review_comment':
        await handlePullRequestComment(event);
        break;

      case 'issues':
        await handleIssue(event);
        break;

      case 'issue_comment':
        await handleIssueComment(event);
        break;

      default:
        console.log(`[GitHub Webhook] Unhandled event type: ${eventType}`);
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('[GitHub Webhook] Error processing webhook:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}

/**
 * Handle pull_request event (opened, synchronize, closed, etc)
 */
async function handlePullRequest(event: any): Promise<void> {
  const { action, pull_request } = event;

  const prTitle = `[PR] ${pull_request.title} (#${pull_request.number})`;
  const prBody = `
**Repository:** ${event.repository.full_name}
**PR:** ${pull_request.html_url}
**Author:** ${pull_request.user.login}
**Status:** ${action}

${pull_request.body || 'No description provided'}
  `.trim();

  console.log(`[GitHub Webhook] PR ${action}: ${prTitle}`);

  // TODO: Create/update thread in Convex
  // await createOrUpdateThread({
  //   title: prTitle,
  //   description: prBody,
  //   source: 'github',
  //   sourceId: pull_request.id,
  // });
}

/**
 * Handle pull_request_review event (submitted, dismissed)
 */
async function handlePullRequestReview(event: any): Promise<void> {
  const { action, review, pull_request } = event;

  const reviewTitle = `[Review] ${pull_request.title} (#${pull_request.number}) - ${review.state}`;
  const reviewBody = `
**Reviewer:** ${review.user.login}
**State:** ${review.state} (${action})
**PR:** ${pull_request.html_url}

${review.body || 'No review comment'}
  `.trim();

  console.log(`[GitHub Webhook] PR review ${action}: ${reviewTitle}`);

  // TODO: Create message in PR thread
  // await addMessageToThread({
  //   threadId: <thread_id>,
  //   body: reviewBody,
  //   author: review.user.login,
  // });
}

/**
 * Handle pull_request_review_comment event
 */
async function handlePullRequestComment(event: any): Promise<void> {
  const { action, comment, pull_request } = event;

  const commentBody = `
**Comment by ${comment.user.login}:**
${comment.body}

[View on GitHub](${comment.html_url})
  `.trim();

  console.log(`[GitHub Webhook] PR comment ${action}`);

  // TODO: Add message to PR thread
}

/**
 * Handle issues event (opened, closed, reopened, labeled)
 */
async function handleIssue(event: any): Promise<void> {
  const { action, issue } = event;

  const issueTitle = `[Issue] ${issue.title} (#${issue.number})`;
  const issueBody = `
**Repository:** ${event.repository.full_name}
**Issue:** ${issue.html_url}
**Author:** ${issue.user.login}
**Status:** ${action}
**Labels:** ${issue.labels.map((l: any) => l.name).join(', ') || 'none'}

${issue.body || 'No description provided'}
  `.trim();

  console.log(`[GitHub Webhook] Issue ${action}: ${issueTitle}`);

  // TODO: Create/update thread in Convex
}

/**
 * Handle issue_comment event
 */
async function handleIssueComment(event: any): Promise<void> {
  const { action, comment, issue } = event;

  const commentBody = `
**Comment by ${comment.user.login}:**
${comment.body}

[View on GitHub](${comment.html_url})
  `.trim();

  console.log(`[GitHub Webhook] Issue comment ${action}`);

  // TODO: Add message to issue thread
}

/**
 * GET /api/github/webhook — health check
 */
export async function GET(request: NextRequest) {
  return NextResponse.json({
    status: 'ok',
    message: 'GitHub webhook endpoint is listening',
  });
}
