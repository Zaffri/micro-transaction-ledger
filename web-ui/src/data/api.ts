const API_GATEWAY = 'http://localhost:8080'
const ACCOUNTS_ENDPOINT = '/accounts'

export interface StatementLine {
  Status: 'settled' | 'rejected_fraud' | 'pending';
  AmountInPennies: number;
  OtherPartyName: string;
  CreatedAt: string;
}

export interface Account {
  ID: number;
  BalanceInPennies: number;
  AccountHolderName: string;
  CreatedAt: string;
  UpdatedAt: string;
  Statement: StatementLine[];
}

export interface PaymentResponse {
  success: boolean;
}

export const getAccounts = async (): Promise<{ data: Account[] } | { error: string }> => {
  try {
    const accountRequests = [
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/1`).then(res => res.json() as Promise<Account>),
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/2`).then(res => res.json() as Promise<Account>),
    ];

    const statementRequests = [
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/1/statement`).then(res => res.json() as Promise<StatementLine[]>),
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/2/statement`).then(res => res.json() as Promise<StatementLine[]>),
    ];
  
    const [accountOne, accountTwo] = await Promise.all(accountRequests);
    const [accountOneStatement, accountTwoStatement] = await Promise.all(statementRequests);
    
    const preparedData: Account[] = [];

    if (accountOne && accountOneStatement) {
      accountOne.Statement = accountOneStatement
      preparedData[0] = accountOne;
    }

    if (accountTwo && accountTwoStatement) {
      accountTwo.Statement = accountTwoStatement
      preparedData[1] = accountTwo;
    }

    return { data: preparedData };
  } catch (err) {
    console.log(err);
    return { error: 'Failed to fetch account data' };
  }
};

export const makePayment = async (sender: number, amount: number): Promise<{ success: boolean } | { error: string }> => {
  if (sender <= 0) {
    // TODO: show err message
    return { error: 'Missing transfer direction: please select transfer details' };
  }

  if (amount <= 0) {
    // TODO: show err message
    return { error: 'Invalid amount: please select a transfer amount greater than 0' };
  }

  try {
    const response = await fetch(
      `${API_GATEWAY}${ACCOUNTS_ENDPOINT}/payment`,
      {
        method: 'POST',
        body: JSON.stringify({
          senderAccountId: sender,
          receiverAccountId: sender == 1 ? 2 : 1,
          amountInPennies: amount
        }),
      },
    );

    const jsonBody = await response.json() as Promise<PaymentResponse>;
    return jsonBody;
  } catch(err) {
    // TODO: show err message
    console.log(err);
    return { error: '' };
  }
};

