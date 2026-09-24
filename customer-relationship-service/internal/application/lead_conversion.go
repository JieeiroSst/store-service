package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type leadConversion struct {
	tx       port.TxRunner
	leads    port.Repository[model.Lead]
	accounts port.Repository[model.Account]
	contacts port.Repository[model.Contact]
	roles    port.Repository[model.AccountContactRole]
	opps     port.Repository[model.Opportunity]
	notifier port.Notifier
}

func NewLeadConversion(
	tx port.TxRunner,
	leads port.Repository[model.Lead],
	accounts port.Repository[model.Account],
	contacts port.Repository[model.Contact],
	roles port.Repository[model.AccountContactRole],
	opps port.Repository[model.Opportunity],
	n port.Notifier,
) port.LeadConverter {
	return &leadConversion{tx: tx, leads: leads, accounts: accounts, contacts: contacts, roles: roles, opps: opps, notifier: n}
}

func (s *leadConversion) Convert(ctx context.Context, leadID uint, in port.ConvertLeadInput) (*port.ConvertLeadResult, error) {
	var res *port.ConvertLeadResult

	err := s.tx.InTx(ctx, func(ctx context.Context) error {
		lead, err := s.leads.GetByID(ctx, leadID)
		if err != nil {
			return err
		}
		if lead.Status == model.LeadConverted {
			return common.ErrInvalidTransition
		}

		fullName := strings.TrimSpace(lead.FirstName + " " + lead.Surname)
		accountName := in.AccountName
		if accountName == "" {
			accountName = fullName
		}

		account := &model.Account{Name: accountName, Phone: lead.Phone}
		if err := s.accounts.Create(ctx, account); err != nil {
			return err
		}
		contact := &model.Contact{AccountID: account.ID, Name: fullName, Email: lead.Email, Phone: lead.Phone}
		if err := s.contacts.Create(ctx, contact); err != nil {
			return err
		}
		if err := s.roles.Create(ctx, &model.AccountContactRole{ContactID: contact.ID, AccountID: account.ID}); err != nil {
			return err
		}

		res = &port.ConvertLeadResult{Lead: lead, Account: account, Contact: contact}
		if in.CreateOpportunity {
			opp := &model.Opportunity{
				AccountID:   account.ID,
				Description: in.OpportunityDescription,
				Amount:      in.OpportunityAmount,
				Stage:       model.StageProspecting,
			}
			if err := s.opps.Create(ctx, opp); err != nil {
				return err
			}
			res.Opportunity = opp
		}

		lead.Status = model.LeadConverted
		lead.ConvertedAccountID, lead.ConvertedContactID = &account.ID, &contact.ID
		return s.leads.Update(ctx, lead.ID, lead)
	})
	if err != nil {
		return nil, err
	}

	notify(ctx, s.notifier, "Lead converted", "Lead #%d became account #%d", res.Lead.ID, res.Account.ID)
	return res, nil
}
