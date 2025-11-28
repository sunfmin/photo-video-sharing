# How to Write Good Specifications
## A Practical Guide to Requirements Documentation

---

# 🎯 What is a Specification?

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

# 🎪 Why User Stories & Acceptance Scenarios?

### The Problem with Traditional Requirements:

❌ "The system shall validate user input"  
- Who is validating? Why? How do we test it?

### The Solution: User-Centered Format

✅ **User Story**: "As a user, I want to upload photos..."  
- Clear **who** and **why**

✅ **Acceptance Scenario**: "Given..., When..., Then..."  
- Clear **how to test**

### Benefits:

- **Shared understanding** between business and QA
- **Direct mapping** to automated tests
- **User focus** prevents building wrong features
- **Testable** from day one

---

# 👥 Who is Your Audience?

Specifications are written for **non-technical stakeholders**:

- Product managers
- Business analysts
- Designers
- Clients/customers
- Project sponsors
- QA testers (before implementation)

**Key Rule**: If a business person can't understand it, it's too technical!

---

# 🎨 The Golden Rule: WHAT vs. HOW

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

# 📖 User Stories: The Foundation

User stories describe **who** needs the feature and **why** it matters.

### Standard Format:

```
As a [user type],
I want to [action/capability],
So that [business value/benefit]
```

### ✅ Good Example (Photo Sharing)

```
As a user, I want to upload my photos and videos 
and be able to view them later, so that I can 
safely store and access my memories from any device.
```

**Notice**: Clear actor, clear action, clear value proposition!

---

# 🎯 Acceptance Scenarios: Given-When-Then

Acceptance scenarios make user stories **testable** using Given-When-Then format.

### Format:

```
Given [initial context/state]
When [action occurs]
Then [expected outcome]
```

### Why This Works:

- **Given** = Sets up the test conditions
- **When** = Describes the user action
- **Then** = Defines success (testable!)

---

# ✅ Acceptance Scenario Examples

### Example 1: Photo Upload

```
US1-AS1: Given I am a user with valid credentials,
         When I select a photo file (JPG, PNG, or HEIC) 
              under 50MB and click upload,
         Then the photo appears in my media gallery 
              within 5 seconds
```

### Example 2: Album Sharing

```
US4-AS5: Given I have an album,
         When I share the entire album with another user,
         Then they can view all current and future items 
              I add to that album
```

**Each scenario = One automated test!**

---

# 🔗 User Stories + Acceptance Scenarios

A complete user story has:

1. **User Story** - The "what" and "why"
2. **Multiple Acceptance Scenarios** - The "how to verify"
3. **Priority** - Business importance (P1, P2, P3)
4. **Independent Test Statement** - Can it be tested alone?

### Example Structure:

```
User Story 1 - Upload Personal Media (Priority: P1)

As a user, I want to upload photos and videos...

Acceptance Scenarios:
- US1-AS1: Upload single photo → appears in gallery
- US1-AS2: Upload video → shows with thumbnail
- US1-AS3: View gallery → see all media by date
- US1-AS4: Click photo → opens full-screen viewer
- US1-AS5: Click video → plays with controls
```

---

# 📝 Writing Testable Requirements

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

# 🎯 Success Criteria: The Gold Standard

Success criteria must be:

1. **Measurable** - Include specific metrics
2. **Technology-agnostic** - No implementation details
3. **User-focused** - Outcomes, not system internals
4. **Verifiable** - Testable without code knowledge

---

# ✅ Success Criteria Examples: The Good (1/2)

```
✓ "Users can complete checkout in under 3 minutes"
  → Measurable: 3 minutes
  → User-focused: checkout completion
  → Verifiable: time the process

✓ "System supports 10,000 concurrent users"
  → Measurable: 10,000 users
  → Business outcome: scalability
  → Verifiable: load testing
```

---

# ✅ Success Criteria Examples: The Good (2/2)

```
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

# ❌ Success Criteria Examples: The Bad (1/2)

```
✗ "API response time is under 200ms"
  → Too technical (API)
  → Better: "Users see results instantly"

✗ "Database can handle 1000 TPS"
  → Implementation detail (database)
  → Better: "System processes 1000 orders per second"
```

---

# ❌ Success Criteria Examples: The Bad (2/2)

```
✗ "React components render efficiently"
  → Framework-specific (React)
  → Better: "Pages load in under 2 seconds"

✗ "Redis cache hit rate above 80%"
  → Technology-specific (Redis)
  → Better: "Frequently accessed data loads instantly"
```

**Tip**: Replace technical terms with user-facing outcomes!

---

# 🤔 Dealing with Uncertainty (1/2)

### Make Informed Guesses (Preferred!)

Use industry standards and common patterns:

- Photo storage limits → 50MB per photo, 500MB per video
- Thumbnail generation → Within 10 seconds
- Session duration → 30 days for remember-me
- Upload batch size → Maximum 50 files at once
- Error handling → User-friendly messages

**Document assumptions** in your spec!

---

# 🤔 Dealing with Uncertainty (2/2)

### Mark for Clarification (Maximum 3!)

Only when:
- ✓ Choice significantly impacts scope
- ✓ Multiple interpretations with different implications
- ✓ No reasonable default exists

### Example:

```markdown
[NEEDS CLARIFICATION: Can users edit photos shared by others, 
or only view them?]
```

**Prioritize**: Scope > Security > UX > Technical details

---

# 🔗 How Everything Connects

### The Complete Picture:

```
User Story (The "Why")
   ↓
Acceptance Scenarios (The "How to Test")
   ↓
Functional Requirements (The "What System Must Do")
   ↓
Success Criteria (The "How to Measure Success")
```

### Example Flow:

1. **US3**: "As a user, I want to share photos with friends"
2. **US3-AS1**: "Given I have a photo, When I share it, Then friend can view it"
3. **FR-019**: "System MUST allow sharing photos by email address"
4. **SC-006**: "Shared media accessible within 10 seconds"

**Each layer adds more detail while staying technology-agnostic!**

---

# 🚫 Common Mistakes to Avoid (1/2)

### 1. Implementation Leakage
❌ "Use S3 and CloudFront for photo storage"  
✅ "System stores photos persistently with fast access"

### 2. Vague Requirements  
❌ "Fast upload times"  
✅ "Photos appear in gallery within 5 seconds"

### 3. Missing Edge Cases
❌ "Users can delete albums"  
✅ "Users can delete albums unless shared; prompts for confirmation"

---

# 🚫 Common Mistakes to Avoid (2/2)

### 4. Subjective Language
❌ "Intuitive sharing experience"  
✅ "90% of users share media successfully on first attempt"

### 5. Untestable Acceptance Scenarios
❌ "When user uploads, then it works properly"  
✅ "When user uploads 50MB photo, then it appears in gallery within 5 seconds"

---

# 📋 Specification Quality Checklist (1/3)

Before considering your spec complete, verify:

### Content Quality
- [ ] No implementation details (languages, frameworks, APIs, databases)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed

---

# 📋 Specification Quality Checklist (2/3)

### User Stories & Scenarios
- [ ] Each user story has "As a... I want... so that..." format
- [ ] All acceptance scenarios use Given-When-Then format
- [ ] Each scenario is independently testable
- [ ] Scenarios are numbered (US#-AS#)
- [ ] Each "Then" statement is specific and measurable

---

# 📋 Specification Quality Checklist (3/3)

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] All requirements testable and unambiguous
- [ ] Success criteria are measurable (include numbers!)
- [ ] Success criteria are technology-agnostic
- [ ] All edge cases identified
- [ ] Functional requirements numbered (FR-###)

---

# 🎓 Real-World Example: Bad → Good

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

**Problems**: 
- All implementation details (React, Node.js, JWT, bcrypt)
- No user story or acceptance scenarios
- Unmeasurable requirements (fast? good UX?)

---

# 🎓 Real-World Example: Bad → Good (cont.)

### ✅ Improved (Good Spec)

```
User Story 2 - User Authentication (Priority: P1)

As a user, I want to create an account and log in securely,
so that my media is private and only accessible to me.

Acceptance Scenarios:

US2-AS1: Given I am a new user,
         When I provide valid email and password (min 8 chars),
         Then my account is created and I am logged in

US2-AS2: Given I have an existing account,
         When I enter my correct email and password,
         Then I am logged into my personal media gallery

US2-AS3: Given I am logged in,
         When I log out and another user logs in,
         Then they see only their own media, not mine

US2-AS4: Given I am not logged in,
         When I try to access upload or gallery features,
         Then I am redirected to the login page
```

---

# 🎓 Real-World Example: Bad → Good (cont. 2)

### ✅ Improved Spec (Continued)

```
Functional Requirements:
- FR-008: System allows account creation with email/password
- FR-009: System enforces minimum password (8 chars, 
          one letter, one number)
- FR-010: System provides secure login authentication
- FR-013: System isolates each user's media

Success Criteria:
- SC-004: 95% of users complete account creation and 
          first upload within 3 minutes
- SC-007: Zero unauthorized access incidents
- SC-008: 90% of users successfully share media on 
          first attempt

Edge Cases:
- What happens when user enters wrong password 3 times?
- What happens when session expires during upload?
```

**Better**: User-focused, testable, technology-agnostic!

---

# 🎯 Key Sections of a Good Spec (1/2)

### Mandatory Sections (Every Feature):

1. **User Stories** - As a [who], I want [what], so that [why]
2. **Acceptance Scenarios** - Given-When-Then test cases (US#-AS#)
3. **Functional Requirements** - Specific, testable "MUST" statements (FR-###)
4. **Success Criteria** - Measurable outcomes (SC-###)
5. **Edge Cases** - What happens when... ?

---

# 🎯 Key Sections of a Good Spec (2/2)

### Optional Sections (Include When Relevant):

6. **Key Entities** - Data objects involved
7. **Scope** - What's included/excluded?
8. **Assumptions** - What are we assuming?
9. **Dependencies** - What does it rely on?

**Tip**: Remove optional sections that don't apply!

---

# 💡 Pro Tips for Spec Writing (1/2)

### 1. One Acceptance Scenario = One Test
Each US#-AS# should map directly to an automated test case.

### 2. Think Like a Tester
Ask: "How would I verify this requirement?" If you can't answer, it's not testable.

### 3. Use the "Explain to Your Boss" Test
If you can't explain it to a non-technical manager, it's too technical.

---

# 💡 Pro Tips for Spec Writing (2/2)

### 4. Be Specific with Numbers
"Fast" → "within 5 seconds"  
"Many users" → "1000 concurrent users"  
"Large files" → "photos up to 50MB, videos up to 500MB"

### 5. Consider Edge Cases Early
"What happens when..." questions reveal gaps in your spec.

### 6. Number Everything
- User Stories: US1, US2, US3...
- Acceptance Scenarios: US1-AS1, US1-AS2...
- Functional Requirements: FR-001, FR-002...
- Success Criteria: SC-001, SC-002...

This makes traceability easy!

---

# ✅ Verification Requirements

Every acceptance scenario should be:

### The Four Pillars of Testability:

1. **Testable** - Can be demonstrated and verified
2. **Complete** - Verifies entire expected behavior
3. **Automated** - Can run repeatedly without manual intervention
4. **Independent** - Can be tested separately from other scenarios

### Example: US1-AS1

```
Given I am a user with valid credentials,
When I select a photo file (JPG, PNG, or HEIC) under 50MB 
     and click upload,
Then the photo appears in my media gallery within 5 seconds
```

**Automated Test**: ✅ Create user → Upload test.jpg (10MB) → 
Verify gallery contains test.jpg → Verify time < 5 seconds

---

# 🔄 The Validation Loop

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

# 📚 Real-World Exercise

### Transform This Bad Requirement:

❌ "Build a photo sharing system using AWS S3, with 
React frontend and GraphQL API. Use Redis for caching 
and ensure sub-200ms response times."

### Your Turn:
- Write a user story (As a... I want... so that...)
- Create 2-3 acceptance scenarios (Given-When-Then)
- Add functional requirements (no tech!)
- Define success criteria (measurable)

---

# 📚 Exercise: Answer

### ✅ Good Version:

```
User Story 3 - Share Media with Users (Priority: P2)

As a user, I want to share specific photos with other users,
so that I can collaborate and share memories with friends.

Acceptance Scenarios:

US3-AS1: Given I have uploaded a photo,
         When I select "Share" and enter another user's email,
         Then that user receives notification and can view 
              the photo in their "Shared with me" section

US3-AS2: Given I have shared a photo with a user,
         When I select "Stop sharing",
         Then that user no longer has access to the photo

US3-AS3: Given another user has shared a photo with me,
         When I view it in my "Shared with me" section,
         Then I can view but not delete or modify it
```

---

# 📚 Exercise: Answer (cont.)

```
Functional Requirements:
- FR-019: System allows sharing individual photos with 
          specific users by email address
- FR-020: System notifies users when media is shared with them
- FR-021: System provides "Shared with me" section
- FR-022: System allows share creators to revoke access anytime
- FR-023: System prevents recipients from deleting/modifying 
          shared content

Success Criteria:
- SC-006: Shared media accessible to recipients within 10 seconds
- SC-008: 90% of users successfully share media on first attempt

Edge Cases:
- What happens when sharing with invalid email?
- What happens when user is removed while viewing shared media?
```

**Notice**: Complete user story with testable scenarios, no technology mentioned!

---

# 🎯 Summary: The Spec Writing Mindset (1/2)

### Always Remember:

1. Start with **User Stories** (As a... I want... so that...)
2. Make them testable with **Acceptance Scenarios** (Given-When-Then)
3. Write **WHAT** the feature does, not **HOW** it's built
4. Write for **business people**, not developers
5. Be **specific and measurable** (use numbers!)
6. Think **user outcomes**, not system internals
7. **Validate** every requirement is testable
8. When unclear, **make informed guesses** (document them!)

---

# 🎯 Summary: The Spec Writing Mindset (2/2)

### The Ultimate Test:

> "Can a non-technical stakeholder understand this spec 
> and know when the feature is done?"

If yes, you've written a good spec! 🎉

---

# ❓ Questions & Discussion (1/2)

**Q: How many acceptance scenarios should each user story have?**  
A: Typically 3-6. Cover the happy path, error cases, and key variations.

**Q: Do I really need Given-When-Then for every scenario?**  
A: Yes! It ensures testability and clarity. If you can't write it, it's not testable.

**Q: How much detail is too much?**  
A: If it describes implementation (code, frameworks, APIs), it's too much.

**Q: What if requirements change?**  
A: Specs are living documents. Update user stories and scenarios as needed.

---

# ❓ Questions & Discussion (2/2)

**Q: How do I handle technical constraints?**  
A: Note them in Dependencies or Assumptions, but keep requirements technology-agnostic.

**Q: What if I need 10 clarifications?**  
A: Make informed guesses for all but the 3 most critical. Document assumptions.

**Q: Can one acceptance scenario test multiple requirements?**  
A: Keep scenarios focused on one main behavior. It's okay to have many small scenarios!

---

# 📖 Resources & Next Steps (1/2)

### Practice Makes Perfect:

1. Review existing specs in your projects
2. Write user stories for each feature
3. Create Given-When-Then acceptance scenarios
4. Identify and remove implementation details
5. Make vague requirements specific and measurable
6. Number everything for traceability
7. Create a quality checklist and validate

---

# 📖 Resources & Next Steps (2/2)

### Key Takeaway:

> "A good specification enables everyone—from 
> business to QA—to know what success looks like, 
> without knowing a single line of code."
>
> "Every acceptance scenario becomes an automated test."

---

# Thank You! 🙏

**Remember**: Great specifications lead to great products!

Questions? Let's discuss!

---

# Appendix: Quick Reference Card (1/3)

## ✅ DO This:

- Start with user stories (As a... I want... so that...)
- Write acceptance scenarios in Given-When-Then format
- Number all elements (US1-AS1, FR-001, SC-001)
- Focus on user value and outcomes
- Use measurable, specific criteria (with numbers!)
- Write for non-technical readers
- Document assumptions
- Consider edge cases ("What happens when...")
- Make every requirement testable

---

# Appendix: Quick Reference Card (2/3)

## ❌ DON'T Do This:

- Mention frameworks, languages, or technologies
- Use vague terms (fast, good, secure, intuitive)
- Assume technical knowledge
- Leave requirements untestable
- Skip acceptance scenarios
- Write "Then it works" (too vague!)
- Over-clarify minor details

## 📏 Quality Bar:

Every requirement must answer:
- Can I test this? (Is it an acceptance scenario?)
- Can a non-developer understand this?
- Is this measuring user outcomes?
- Is this free of implementation details?
- Does it include specific numbers/thresholds?

**If all YES → Good requirement!** ✨

---

# Appendix: Quick Reference Card (3/3)

## 📋 Quick Format Reference:

**User Story**: As a [user], I want [action], so that [value]

**Acceptance Scenario**: Given [context], When [action], Then [outcome]

**Functional Requirement**: System MUST [specific capability]

**Success Criterion**: [X]% of users [achieve outcome] in [Y time]

**Edge Case**: What happens when [unusual situation]?

