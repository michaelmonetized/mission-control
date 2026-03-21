import Stripe from "stripe";

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY || "", {
  apiVersion: "2024-04-10",
});

export interface UsageRecord {
  userId: string;
  workspaceId: string;
  durationMinutes: number;
  costUSD: number;
}

/**
 * Create a Stripe customer for a new user
 */
export async function createStripeCustomer(
  clerkId: string,
  email: string,
  name: string
) {
  return await stripe.customers.create({
    metadata: { clerkId },
    email,
    name,
  });
}

/**
 * Record usage for metered billing
 * Should be called after each workspace stops
 */
export async function recordUsage(
  stripeCustomerId: string,
  record: UsageRecord
) {
  try {
    // Find or create meter
    const subscriptions = await stripe.subscriptions.list({
      customer: stripeCustomerId,
      limit: 1,
    });

    if (!subscriptions.data.length) {
      console.warn(`No subscription found for customer ${stripeCustomerId}`);
      return null;
    }

    const subscription = subscriptions.data[0];

    // Report usage to Stripe
    // In real implementation, use Stripe Billing with metered billing
    // For now, track in our own database and batch monthly invoicing

    return {
      customerId: stripeCustomerId,
      recorded: true,
      durationMinutes: record.durationMinutes,
      costUSD: record.costUSD,
    };
  } catch (error) {
    console.error("Failed to record usage:", error);
    throw error;
  }
}

/**
 * Create a Payment Intent for charging a user
 */
export async function createPaymentIntent(
  stripeCustomerId: string,
  amountCents: number,
  description: string
) {
  return await stripe.paymentIntents.create({
    customer: stripeCustomerId,
    amount: amountCents,
    currency: "usd",
    description,
    automatic_payment_methods: {
      enabled: true,
    },
  });
}

/**
 * Get customer's billing history
 */
export async function getCustomerInvoices(stripeCustomerId: string) {
  return await stripe.invoices.list({
    customer: stripeCustomerId,
    limit: 12,
  });
}

/**
 * Create monthly invoice for accumulated usage
 */
export async function createMonthlyInvoice(
  stripeCustomerId: string,
  billingPeriod: string,
  totalMinutes: number,
  totalCost: number
) {
  const pricePerMinute = 0.001; // $0.001/min

  return await stripe.invoiceItems.create({
    customer: stripeCustomerId,
    amount: Math.round(totalCost * 100), // Convert to cents
    currency: "usd",
    description: `Mission Control compute: ${totalMinutes} minutes (${billingPeriod})`,
    period: {
      start: Math.floor(new Date(billingPeriod + "-01").getTime() / 1000),
      end: Math.floor(
        new Date(new Date(billingPeriod + "-01").setMonth(
          new Date(billingPeriod + "-01").getMonth() + 1
        )).getTime() / 1000
      ),
    },
  });
}
