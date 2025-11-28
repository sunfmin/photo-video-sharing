# How to Write Good Specifications
## A Practical Guide to Requirements Documentation

---

## 🎯 What is a Specification?

A **specification** (spec) is a document that describes:
- **WHAT** the feature should do
- **WHY** it matters to users/business
- **WHO** will use it
- **WHEN** it's considered successful

### What a Spec is NOT:
- ❌ Technical implementation details
- ❌ Code architecture decisions
- ❌ Technology choices (languages, frameworks, databases)
- ❌ API designs or database schemas

---

## 👥 Who is Your Audience?

Specifications are written for **non-technical stakeholders**:

- Product managers
- Business analysts
- Designers
- Clients/customers
- Project sponsors
- QA testers (before implementation)

**Key Rule**: If a business person can't understand it, it's too technical!

---

## 🎨 The Golden Rule: WHAT vs. HOW

### ✅ WHAT (Specification Language)

- "Users can upload photos"
- "The system validates email addresses"
- "Customers receive order confirmation"
- "Data is kept secure"

### ❌ HOW (Implementation Language)

- "React component for file upload"
- "Use regex pattern for email validation"
- "Send email via SendGrid API"
- "Store passwords with bcrypt encryption"

**Remember**: Focus on USER OUTCOMES, not system internals!

---

## 📝 Writing Testable Requirements

Every requirement must be **testable and unambiguous**.

### ✅ Good Requirements

- "Users can upload images up to 10MB"
- "Search results appear within 2 seconds"
- "Users can share albums with up to 50 people"
- "The system prevents duplicate email registrations"

### ❌ Bad Requirements

- "Users can upload files quickly" (vague: how quickly?)
- "The system is fast" (unmeasurable)
- "Users have a good experience" (subjective)
- "It should be secure" (undefined: what does secure mean?)

**Test**: Can you verify this requirement without knowing the code?

---

## 🎯 Success Criteria: The Gold Standard

Success criteria must be:

1. **Measurable** - Include specific metrics
2. **Technology-agnostic** - No implementation details
3. **User-focused** - Outcomes, not system internals
4. **Verifiable** - Testable without code knowledge

---

## ✅ Success Criteria Examples: The Good

```
✓ "Users can complete checkout in under 3 minutes"
  → Measurable: 3 minutes
  → User-focused: checkout completion
  → Verifiable: time the process

✓ "System supports 10,000 concurrent users"
  → Measurable: 10,000 users
  → Business outcome: scalability
  → Verifiable: load testing

✓ "95% of searches return results in under 1 second"
  → Measurable: 95%, 1 second
  → User experience: fast search
  → Verifiable: performance monitoring

✓ "Task completion rate improves by 40%"
  → Measurable: 40% improvement
  → Business value: efficiency
  → Verifiable: analytics comparison
```

---

## ❌ Success Criteria Examples: The Bad

```
✗ "API response time is under 200ms"
  → Too technical (API)
  → Better: "Users see results instantly"

✗ "Database can handle 1000 TPS"
  → Implementation detail (database)
  → Better: "System processes 1000 orders per second"

✗ "React components render efficiently"
  → Framework-specific (React)
  → Better: "Pages load in under 2 seconds"

✗ "Redis cache hit rate above 80%"
  → Technology-specific (Redis)
  → Better: "Frequently accessed data loads instantly"
```

**Tip**: Replace technical terms with user-facing outcomes!

---

## 🤔 Dealing with Uncertainty

When details are unclear, you have two choices:

### 1. Make Informed Guesses (Preferred!)

Use industry standards and common patterns:
- Data retention → Standard practices for the domain
- Performance → Web/mobile app expectations
- Authentication → OAuth2 or session-based for web
- Error handling → User-friendly messages

**Document assumptions** in your spec!

### 2. Mark for Clarification (Maximum 3!)

Only when:
- ✓ Choice significantly impacts scope
- ✓ Multiple interpretations with different implications
- ✓ No reasonable default exists

```markdown
[NEEDS CLARIFICATION: Can users edit shared albums, 
or only view them?]
```

**Prioritize**: Scope > Security > UX > Technical details

---

## 🚫 Common Mistakes to Avoid

### 1. Implementation Leakage
❌ "Use PostgreSQL for data storage"  
✅ "System stores user data persistently"

### 2. Vague Requirements
❌ "Fast response times"  
✅ "Pages load within 2 seconds"

### 3. Missing Edge Cases
❌ "Users can delete albums"  
✅ "Users can delete albums unless shared with others"

### 4. Subjective Language
❌ "Good user experience"  
✅ "90% of users complete tasks on first attempt"

---

## 📋 Specification Quality Checklist

Before considering your spec complete, verify:

### Content Quality
- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous
- [ ] Success criteria are measurable
- [ ] Success criteria are technology-agnostic
- [ ] All edge cases identified
- [ ] Scope is clearly bounded

---

## 🎓 Real-World Example: Bad → Good

### ❌ Original (Poor Spec)
```
Feature: User System

Build a user authentication system with React 
and Node.js. Use JWT tokens and bcrypt for 
passwords. The API should be RESTful.

Requirements:
- Fast login
- Secure storage
- Good UX
```

**Problems**: All implementation details, unmeasurable requirements!

---

## 🎓 Real-World Example: Bad → Good (cont.)

### ✅ Improved (Good Spec)

```
Feature: User Account Management

Purpose: Allow users to create accounts and 
securely access their personal photo libraries.

Requirements:
- Users can register with email and password
- Email addresses must be unique and valid
- Passwords must be at least 12 characters
- Users remain logged in for 30 days
- Login process completes in under 3 seconds

Success Criteria:
- 95% of users complete registration in under 2 minutes
- Zero unauthorized account access incidents
- Account recovery success rate above 90%

Assumptions:
- Users have valid email addresses
- Standard web security practices are sufficient
```

**Better**: Clear, measurable, technology-agnostic!

---

## 🎯 Key Sections of a Good Spec

1. **Overview** - What is this feature?
2. **Purpose** - Why does it matter?
3. **User Scenarios** - Who uses it and how?
4. **Functional Requirements** - What must it do?
5. **Success Criteria** - How do we know it works?
6. **Scope** - What's included/excluded?
7. **Assumptions** - What are we assuming?
8. **Edge Cases** - What can go wrong?
9. **Dependencies** - What does it rely on?

**Tip**: Remove optional sections that don't apply!

---

## 💡 Pro Tips for Spec Writing

### 1. Think Like a Tester
Ask: "How would I verify this requirement?"

### 2. Use the "Explain to Your Boss" Test
If you can't explain it to a non-technical manager, it's too technical.

### 3. Document Your Assumptions
Don't leave implicit decisions unstated.

### 4. Be Specific with Numbers
"Fast" → "under 2 seconds"  
"Many users" → "10,000 concurrent users"  
"Large files" → "up to 100MB"

### 5. Consider Edge Cases Early
What happens when things go wrong?

---

## 🔄 The Validation Loop

```
1. Write initial spec
   ↓
2. Check against quality criteria
   ↓
3. Issues found?
   ├─ Yes → Fix and return to step 2
   └─ No → Spec is ready!
```

**Maximum 3 iterations** before escalating issues.

---

## 📚 Real-World Exercise

### Transform This Bad Requirement:

❌ "Build a photo sharing system using AWS S3, with 
React frontend and GraphQL API. Use Redis for caching 
and ensure sub-200ms response times."

### Your Turn:
- Remove implementation details
- Focus on user outcomes
- Make it measurable
- Keep it technology-agnostic

---

## 📚 Exercise: Answer

### ✅ Good Version:

```
Feature: Photo Sharing

Purpose: Enable users to share their photos 
with friends and family.

Requirements:
- Users can upload photos up to 10MB each
- Users can organize photos into albums
- Users can share albums via link or email
- Shared albums load within 2 seconds
- Photos remain accessible for 10 years

Success Criteria:
- Users share albums in under 1 minute
- 95% of shared links work on first access
- Photo viewing experience is smooth (no delays)
```

**Notice**: No technology mentioned, all measurable!

---

## 🎯 Summary: The Spec Writing Mindset

### Always Remember:

1. **WHAT** the feature does, not **HOW** it's built
2. Write for **business people**, not developers
3. Be **specific and measurable**
4. Think **user outcomes**, not system internals
5. **Validate** before moving to implementation
6. When unclear, **make informed guesses** (document them!)
7. **Limit clarifications** to critical decisions only

### The Ultimate Test:

> "Can a non-technical stakeholder understand this spec 
> and know when the feature is done?"

If yes, you've written a good spec! 🎉

---

## ❓ Questions & Discussion

### Common Questions:

**Q: How much detail is too much?**  
A: If it describes implementation, it's too much.

**Q: What if requirements change?**  
A: Specs should be living documents. Update them!

**Q: How do I handle technical constraints?**  
A: Note them in Dependencies or Assumptions sections.

**Q: What if I need 10 clarifications?**  
A: Make informed guesses for all but the 3 most critical.

---

## 📖 Resources & Next Steps

### Practice Makes Perfect:

1. Review existing specs in your projects
2. Identify implementation details to remove
3. Make vague requirements specific
4. Add measurable success criteria
5. Create a quality checklist

### Key Takeaway:

> "A good specification enables everyone—from 
> business to QA—to know what success looks like, 
> without knowing a single line of code."

---

## Thank You! 🙏

**Remember**: Great specifications lead to great products!

Questions? Let's discuss!

---

# Appendix: Quick Reference Card

## ✅ DO This:

- Focus on user value and outcomes
- Use measurable, specific criteria
- Write for non-technical readers
- Document assumptions
- Test requirements for clarity
- Consider edge cases

## ❌ DON'T Do This:

- Mention frameworks or languages
- Use vague terms (fast, good, secure)
- Assume technical knowledge
- Leave requirements untestable
- Skip validation steps
- Over-clarify minor details

## 📏 Quality Bar:

Every requirement must answer:
- Can I test this?
- Can a non-developer understand this?
- Is this measuring user outcomes?
- Is this free of implementation details?

**If all YES → Good requirement!** ✨

